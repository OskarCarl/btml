package peer

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"time"

	"github.com/quic-go/quic-go"
	"github.com/vs-ude/btml/internal/model"
	"google.golang.org/protobuf/proto"
)

func (me *Me) Listen() {
	defer func() {
		me.server.Close()
		me.Wg.Done()
	}()

	listener, err := me.server.Listen(me.tlsConfig, generateQUICConfig())
	if err != nil {
		log.Default().Printf("Error listening: %v", err)
		return
	}

	for {
		conn, err := listener.Accept(me.Ctx)
		if err != nil {
			if me.Ctx.Err() != nil {
				log.Default().Print("Stopping the listener")
				return
			}
			log.Default().Printf("Error accepting connection: %v", err)
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
		stream, err := conn.AcceptStream(context.Background())
		if err != nil {
			log.Default().Printf("Error accepting stream: %v", err)
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
				log.Default().Printf("Error reading message length: %v", err)
			}
			return
		}

		msgLen := binary.BigEndian.Uint32(lenBuf)

		// Read the actual message
		msgBuf := make([]byte, msgLen)
		_, err = io.ReadFull(stream, msgBuf)
		if err != nil {
			log.Default().Printf("Error reading message body: %v", err)
			return
		}

		// Unmarshal the protobuf message
		update := &ModelUpdate{}
		err = proto.Unmarshal(msgBuf, update)
		if err != nil {
			log.Default().Printf("Error unmarshaling model update: %v", err)
			continue
		}

		w := model.NewSimpleWeights(update.Weights)

		log.Default().Printf("Received model update from %s, age: %d, weights size: %d",
			update.Source, update.Age, len(update.Weights))
		me.data.incomingChan <- w
	}
}

func (me *Me) Outgoing() {
	defer me.Wg.Done()

	for {
		select {
		case <-me.Ctx.Done():
			return
		case data := <-me.data.outgoingChan:
			me.data.outgoingStorage[data.GetAge()] = data
			for name, peer := range me.peerset.Active {
				go me.sendPeer(data, peer, name)
			}
		}
	}
}

func (me *Me) sendPeer(data model.Weights, peer *KnownPeer, name string) {
	log.Default().Printf("Connecting to peer %s", name)
	conn, err := me.getOrEstablishConnection(peer)
	if err != nil {
		log.Default().Printf("Failed to establish connection to %s: %v", name, err)
		return
	}
	log.Default().Printf("Connected to peer %s at %s", name, peer.P.Addr.String())

	// Open a new stream for sending data
	stream, err := conn.OpenStreamSync(context.Background())
	if err != nil {
		log.Default().Printf("Failed to open stream to %s: %v", name, err)
		return
	}

	// Create and marshal the model update
	update := &ModelUpdate{
		Source:  me.config.Name,
		Weights: data.Get(),
		Age:     int64(data.GetAge()),
	}

	msgBytes, err := proto.Marshal(update)
	if err != nil {
		log.Default().Printf("Error marshaling model update for %s: %v", name, err)
		return
	}

	// Write length prefix
	lenBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(lenBuf, uint32(len(msgBytes)))
	_, err = stream.Write(lenBuf)
	if err != nil {
		log.Default().Printf("Error writing message length to %s: %v", name, err)
		return
	}

	// Write the actual message
	log.Default().Printf("Sending model update to %s with age %d", name, update.Age)
	_, err = stream.Write(msgBytes)
	if err != nil {
		log.Default().Printf("Error sending model update to %s: %v", name, err)
	} else {
		peer.MostRecentUpdate = data.GetAge()
	}
	stream.Close()
}

func (me *Me) getOrEstablishConnection(peer *KnownPeer) (quic.Connection, error) {
	connI, ok := me.conns.Load(peer.P.Addr.String())
	if !ok {
		me.conns.Store(peer.P.Addr.String(), nil)
		conn, err := me.dialPeer(peer.P.Addr)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to %s: %v", peer.P.Addr.String(), err)
		}
		me.conns.Store(peer.P.Addr.String(), conn)
		connI = conn
	} else if connI == nil {
		time.Sleep(time.Second * 4)
		connI, ok = me.conns.Load(peer.P.Addr.String())
		if !ok || connI == nil {
			return nil, fmt.Errorf("connection establishment started but not completed")
		}
	}
	return connI.(quic.Connection), nil
}

func (me *Me) dialPeer(addr net.Addr) (quic.Connection, error) {
	return me.server.Dial(context.Background(), addr, me.tlsConfig, me.quicConfig)
}
