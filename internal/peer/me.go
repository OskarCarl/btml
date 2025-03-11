package peer

import (
	"context"
	"crypto/tls"
	"log"
	"net"
	"sync"
	"time"

	"github.com/quic-go/quic-go"
	"github.com/vs-ude/btml/internal/model"
)

type storage struct {
	incomingChan    chan model.Weights
	outgoingChan    chan model.Weights
	outgoingStorage map[int]model.Weights
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
	conns      sync.Map // map[string]quic.Connection
	data       storage
}

func NewMe(config *Config) *Me {
	ctx, cancel := context.WithCancel(context.Background())
	return &Me{
		Wg:     sync.WaitGroup{},
		Ctx:    ctx,
		cancel: cancel,
		config: config,
		quicConfig: &quic.Config{
			KeepAlivePeriod: time.Second * 10,
			MaxIdleTimeout:  time.Second * 60,
		},
		conns:     sync.Map{},
		tlsConfig: generateTLSConfig(),
		data: storage{
			incomingChan:    make(chan model.Weights, 10),
			outgoingChan:    make(chan model.Weights, 5),
			outgoingStorage: make(map[int]model.Weights),
		},
	}
}

func (me *Me) Setup() {
	addr, err := net.ResolveUDPAddr("udp", ":0")
	if err != nil {
		log.Default().Panicf("Error resolving UDP address: %v\n", err)
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Default().Panicf("Error listening on UDP address: %v\n", err)
	}
	listener := &quic.Transport{
		Conn: conn,
	}
	me.server = listener
	me.localAddr = conn.LocalAddr()
	log.Default().Printf("QUIC listener started on %s", me.localAddr.String())
}

func (me *Me) Shutdown() {
	me.cancel()
	me.tracker.Leave()
	me.Wg.Wait()
}
