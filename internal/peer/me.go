package peer

import (
	"context"
	"crypto/tls"
	"errors"
	"log/slog"
	"net"
	"sync"

	"github.com/quic-go/quic-go"
	"github.com/vs-ude/btml/internal/model"
	"github.com/vs-ude/btml/internal/telemetry"
)

type storage struct {
	incomingChan    chan *model.Weights
	outgoingChan    chan *model.Weights
	outgoingStorage map[int]*model.Weights
	incMutex        sync.Mutex
}

// Me is the peer we use
type Me struct {
	Wg         sync.WaitGroup
	Ctx        context.Context
	cancel     context.CancelFunc
	config     *Config
	quicConfig *quic.Config
	localAddr  net.Addr
	server     *quic.Transport
	tlsConfig  *tls.Config
	tracker    *Tracker
	peerset    *PeerSet
	pss        PeerSelectionStrategy
	pds        DistributionStrategy
	conns      sync.Map // map[string]quic.Connection
	data       storage
	telemetry  *telemetry.Client
}

func NewMe(config *Config, telemetry *telemetry.Client) *Me {
	ctx, cancel := context.WithCancel(context.Background())
	return &Me{
		Wg:         sync.WaitGroup{},
		Ctx:        ctx,
		cancel:     cancel,
		config:     config,
		pss:        &RandomPeerSelectionStrategy{},
		pds:        NewQuadraticDistribution(10, 40),
		quicConfig: generateQUICConfig(),
		conns:      sync.Map{},
		tlsConfig:  generateTLSConfig(),
		data: storage{
			incomingChan:    make(chan *model.Weights, 10),
			outgoingChan:    make(chan *model.Weights, 5),
			outgoingStorage: make(map[int]*model.Weights),
		},
		telemetry: telemetry,
	}
}

func (me *Me) Setup() {
	addr, err := net.ResolveUDPAddr("udp", me.config.Addr+":0")
	if err != nil {
		slog.Error("Failed resolving UDP address", "error", err)
		panic(err)
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		slog.Error("Failed listening on UDP address", "error", err)
		panic(err)
	}
	listener := &quic.Transport{
		Conn: conn,
	}
	me.server = listener
	me.localAddr = conn.LocalAddr()
	slog.Info("QUIC listener started", "addr", me.localAddr.String())
}

func (me *Me) Send(w *model.Weights) {
	me.data.outgoingChan <- w
}

// ListenForWeights listens for incoming weights and returns a channel to
// receive them. Can only be called once, subsequent calls will return an error.
func (me *Me) ListenForWeights() (<-chan *model.Weights, error) {
	ok := me.data.incMutex.TryLock()
	if !ok {
		return nil, errors.New("someone is already listening for incoming weights")
	}
	return me.data.incomingChan, nil
}

func (me *Me) Shutdown() {
	me.cancel()
	me.tracker.Leave()
	close(me.data.incomingChan)
	me.Wg.Wait()
}
