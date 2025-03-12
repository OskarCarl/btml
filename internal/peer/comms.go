package peer

import (
	"encoding/binary"
	"errors"
	"io"
	"log/slog"
	"net"
	"sync"

	"github.com/quic-go/quic-go"
	"github.com/vs-ude/btml/internal/model"
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

		me.conns.Store(conn.RemoteAddr().String(), conn)
		go me.handleConnection(conn)
	}
}

func (me *Me) handleConnection(conn quic.Connection) {
	defer func() {
		conn.CloseWithError(0, "closed")
		me.conns.Delete(conn.RemoteAddr().String())
	}()

	for {
		stream, err := conn.AcceptStream(me.Ctx)
		if err != nil {
			slog.Warn("Failed accepting stream", "error", err)
			return
		}

		go me.handleStream(stream)
	}
}

func (me *Me) handleStream(stream quic.Stream) {
	defer stream.Close()

	for {
		// Read message length prefix (4 bytes)
		lenBuf := make([]byte, 4)
		_, err := io.ReadFull(stream, lenBuf)
		if err != nil {
			if !errors.Is(err, io.EOF) {
				slog.Warn("Failed reading message length", "error", err)
			}
			return
		}

		msgLen := binary.BigEndian.Uint32(lenBuf)

		// Read the actual message
		msgBuf := make([]byte, msgLen)
		_, err = io.ReadFull(stream, msgBuf)
		if err != nil {
			slog.Warn("Failed reading message body", "error", err)
			return
		}

		// Unmarshal the protobuf message
		update := &ModelUpdate{}
		err = proto.Unmarshal(msgBuf, update)
		if err != nil {
			slog.Warn("Failed unmarshaling model update", "error", err)
			continue
		}

		w := model.NewWeights(update.Weights)

		slog.Info("Received model update", "source", update.Source, "age", update.Age)
		me.data.incomingChan <- w
	}
}

func (me *Me) Outgoing() {
	defer me.Wg.Done()

	wg := &sync.WaitGroup{}
	for {
		select {
		case <-me.Ctx.Done():
			return
		case data := <-me.data.outgoingChan:
			wg.Wait() // We wait here so the application can be stopped at any time
			me.data.outgoingStorage[data.GetAge()] = data
			bytes, err := marshalUpdate(data, me.config.Name)
			if err != nil {
				slog.Warn("Failed marshaling model update", "error", err)
				continue
			}
			for _, peer := range me.peerset.Active {
				wg.Add(1)
				go peer.Send(bytes, data.GetAge(), wg, me.Ctx, me.dialPeer)
			}
		}
	}
}

func marshalUpdate(data *model.Weights, source string) ([]byte, error) {
	// Create and marshal the model update
	update := &ModelUpdate{
		Source:  source,
		Weights: data.Get(),
		Age:     int64(data.GetAge()),
	}

	return proto.Marshal(update)
}

func (me *Me) dialPeer(addr net.Addr) (quic.Connection, error) {
	return me.server.Dial(me.Ctx, addr, me.tlsConfig, me.quicConfig)
}
