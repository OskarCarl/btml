package peer

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"time"

	"github.com/quic-go/quic-go"
	"github.com/vs-ude/btml/internal/model"
	"github.com/vs-ude/btml/internal/structs"
	"google.golang.org/protobuf/proto"
)

func (me *Me) Listen() {
	defer func() {
		me.server.Close()
		me.Wg.Done()
	}()

	listener, err := me.server.Listen(me.tlsConfig, me.quicConfig)
	if err != nil {
		slog.Error("Listening failed", "error", err)
		return
	}

	for {
		conn, err := listener.Accept(me.Ctx)
		if err != nil {
			if me.Ctx.Err() != nil {
				slog.Info("Stopping the listener")
				return
			}
			slog.Warn("Failed accepting connection", "error", err)
			continue
		}
		if me.peerset.Space() < 1 {
			conn.CloseWithError(quic.ApplicationErrorCode(CHOKED), "peer set full")
			continue
		}

		go me.handleConnection(conn)
	}
}

func (me *Me) handleConnection(conn *quic.Conn) {
	stream, err := conn.AcceptStream(me.Ctx)
	if err != nil {
		if qerr, ok := err.(*quic.ApplicationError); !ok || qerr.ErrorCode != quic.ApplicationErrorCode(CHOKED) {
			slog.Warn("Failed accepting stream", "error", err)
		}
		conn.CloseWithError(0, "closed")
		return
	}
	err = me.handlePeerInfo(stream, conn.RemoteAddr().(*net.UDPAddr))
	if err != nil {
		slog.Warn("New connection but not an active peer", "error", err)
		conn.CloseWithError(quic.ApplicationErrorCode(CHOKED), "closed")
		return
	}

	defer func() {
		conn.CloseWithError(0, "closed")
	}()
	if me.telemetry != nil {
		me.telemetry.RecordActivePeers(me.peerset.UnchokedToString())
	}
	for {
		stream, err = conn.AcceptStream(me.Ctx)
		if err != nil {
			if qerr, ok := err.(*quic.ApplicationError); !ok || qerr.ErrorCode != quic.ApplicationErrorCode(CHOKED) {
				slog.Warn("Failed accepting stream", "error", err)
			}
			return
		}

		go me.handleStream(stream)
	}
}

func (me *Me) handlePeerInfo(stream *quic.Stream, addr *net.UDPAddr) error {
	defer stream.Close()
	msgLen, err := readLengthPrefix(stream)
	if err != nil {
		return fmt.Errorf("Unable to read PeerInfo length %w", err)
	}
	msgBuf := make([]byte, msgLen)
	_, err = io.ReadFull(stream, msgBuf)
	if err != nil {
		return fmt.Errorf("Unable to read PeerInfo body %w", err)
	}
	peerInfo := &PeerInfo{}
	err = proto.Unmarshal(msgBuf, peerInfo)
	if err != nil {
		return fmt.Errorf("Unable to unmarshal PeerInfo %w", err)
	}

	p := &structs.Peer{
		Name:        peerInfo.Id,
		Fingerprint: peerInfo.Fingerprint,
		Addr:        addr,
		LastSeen:    time.Now(),
	}
	return me.peerset.Add(p)
}

func (me *Me) handleStream(stream *quic.Stream) {
	defer stream.Close()

	for {
		msgLen, err := readLengthPrefix(stream)
		if err != nil {
			if qerr, ok := err.(*quic.ApplicationError); !errors.Is(err, io.EOF) && (!ok || qerr.ErrorCode != quic.ApplicationErrorCode(CHOKED)) {
				slog.Warn("Failed reading message length", "error", err)
			}
			return
		}

		// Read the actual message
		msgBuf := make([]byte, msgLen)
		_, err = io.ReadFull(stream, msgBuf)
		if err != nil {
			if qerr, ok := err.(*quic.ApplicationError); !ok || qerr.ErrorCode != quic.ApplicationErrorCode(CHOKED) {
				slog.Warn("Failed reading message body", "error", err)
			}
			return
		}

		// Unmarshal the protobuf message
		update := &ModelUpdate{}
		err = proto.Unmarshal(msgBuf, update)
		if err != nil {
			slog.Warn("Failed unmarshaling model update", "error", err)
			continue
		}

		w := structs.NewWeights(update.Weights, int(update.Age))

		slog.Info("Received model update", "source", update.Source, "age", update.Age)
		me.data.incomingChan <- model.NewWeightsWithCallback(w, me.peerset.known[update.GetSource()].UpdateScore)
	}
}

// readLengthPrefix extracts the message length prefix (4 bytes)
func readLengthPrefix(stream *quic.Stream) (uint32, error) {
	lenBuf := make([]byte, 4)
	_, err := io.ReadFull(stream, lenBuf)
	if err != nil {
		return 0, err
	}

	return binary.BigEndian.Uint32(lenBuf), nil
}
