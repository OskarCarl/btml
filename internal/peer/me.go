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
	"github.com/vs-ude/btml/internal/structs"
	"github.com/vs-ude/btml/internal/telemetry"
	"google.golang.org/protobuf/proto"
)

type storage struct {
	incomingChan    chan *model.WeightsWithCallback
	outgoingChan    chan *structs.Weights
	outgoingStorage map[int]*structs.Weights
	incMutex        sync.Mutex
}

var myPeerInfo []byte

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
	pds        StorageStrategy
	data       storage
	telemetry  *telemetry.Client
}

func NewMe(config *Config, telemetry *telemetry.Client, p *structs.Peer) *Me {
	ctx, cancel := context.WithCancel(context.Background())
	myPeerInfo, _ = proto.Marshal(&PeerInfo{
		Id:          p.Name,
		Fingerprint: p.Fingerprint,
	})
	return &Me{
		Wg:         sync.WaitGroup{},
		Ctx:        ctx,
		cancel:     cancel,
		config:     config,
		pss:        &RandomPeerSelectionStrategy{},
		pds:        NewDoubleAgeStorage(10, 40),
		quicConfig: generateQUICConfig(),
		tlsConfig:  generateTLSConfig(),
		data: storage{
			incomingChan:    make(chan *model.WeightsWithCallback, 10),
			outgoingChan:    make(chan *structs.Weights, 5),
			outgoingStorage: make(map[int]*structs.Weights),
		},
		telemetry: telemetry,
	}
}

func (me *Me) Setup() {
	_, port, err := net.SplitHostPort(me.config.Addr)
	if port == "" {
		me.config.Addr += ":0"
	}
	addr, err := net.ResolveUDPAddr("udp", me.config.Addr)
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

func (me *Me) Send(w *structs.Weights) {
	me.data.outgoingChan <- w
}

// ListenForWeights listens for incoming weights and returns a channel to
// receive them. Can only be called once, subsequent calls will return an error.
func (me *Me) ListenForWeights() (<-chan *model.WeightsWithCallback, error) {
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

func (me *Me) ManualPeerSet(ps *PeerSet) {
	me.peerset = ps
}
