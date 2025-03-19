package peer

import (
	"context"
	"encoding/binary"
	"fmt"
	"log/slog"
	"net"
	"sync"

	"github.com/quic-go/quic-go"
	"github.com/vs-ude/btml/internal/structs"
	"github.com/vs-ude/btml/internal/telemetry"
	"github.com/vs-ude/btml/internal/trust"
)

type peerStatus int

const (
	ERR peerStatus = iota
	UNCHOKED
	CHOKED
	UNKNOWN
)

type KnownPeer struct {
	S              trust.Score
	LastUpdatedAge int
	State          peerStatus
	conn           quic.Connection
	telemetry      *telemetry.Client
	structs.Peer
	sync.Mutex
}

func NewKnownPeer(p *structs.Peer, telemetry *telemetry.Client) *KnownPeer {
	return &KnownPeer{
		S:              0,
		LastUpdatedAge: 0,
		State:          CHOKED,
		conn:           nil,
		telemetry:      telemetry,
		Peer:           *p.Copy(),
	}
}

func (kp *KnownPeer) Update(p *structs.Peer) {
	if !(kp.Addr.IP.Equal(p.Addr.IP) && kp.Addr.Port == p.Addr.Port) {
		kp.closeConn("addr change")
		kp.Addr = p.Addr
	}
}

func (kp *KnownPeer) closeConn(reason string) {
	if kp.conn != nil {
		kp.Lock()
		defer kp.Unlock()
		kp.conn.CloseWithError(quic.ApplicationErrorCode(quic.NoError), reason)
		kp.conn = nil
	}
}

func (kp *KnownPeer) unchoke() error {
	kp.State = UNCHOKED
	return nil
}

func (kp *KnownPeer) choke() error {
	kp.State = CHOKED
	kp.closeConn("peer choked")
	return nil
}

func (kp *KnownPeer) Send(data []byte, age int, wg *sync.WaitGroup, ctx context.Context, dial func(addr net.Addr) (quic.Connection, error)) {
	defer wg.Done()
	conn, err := kp.getOrEstablishConnection(dial)
	if err != nil {
		slog.Warn("Failed to establish connection", "peer", kp.Name, "error", err)
		return
	}

	// Open a new stream for sending data
	stream, err := conn.OpenStreamSync(ctx)
	if err != nil {
		slog.Warn("Failed to open stream", "peer", kp.Name, "error", err)
		return
	}
	defer stream.Close()

	// Write length prefix
	lenBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(lenBuf, uint32(len(data)))
	_, err = stream.Write(lenBuf)
	if err != nil {
		slog.Warn("Failed writing message length", "peer", kp.Name, "error", err)
		return
	}

	// Write the actual message
	slog.Info("Sending data", "peer", kp.Name)
	_, err = stream.Write(data)
	if err != nil {
		slog.Warn("Failed sending data", "peer", kp.Name, "error", err)
	} else {
		kp.LastUpdatedAge = age
		if kp.telemetry != nil {
			kp.telemetry.RecordSend(age, kp.Name)
		}
	}
}

func (kp *KnownPeer) getOrEstablishConnection(dial func(addr net.Addr) (quic.Connection, error)) (quic.Connection, error) {
	if kp.conn == nil {
		kp.Lock()
		defer kp.Unlock()

		slog.Debug("Connecting to peer", "peer", kp.Name)
		conn, err := dial(kp.Addr)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to %s: %v", kp.Addr.String(), err)
		}
		kp.conn = conn
	}
	return kp.conn, nil
}
