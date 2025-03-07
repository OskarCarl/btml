package peer

import (
	"context"
	"errors"
	"io"
	"log"
	"net"

	"github.com/quic-go/quic-go"
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
		buf := make([]byte, 1024)
		n, err := stream.Read(buf)
		if err != nil && !errors.Is(err, io.EOF) {
			log.Default().Printf("Error reading from stream: %v", err)
			return
		}
		log.Default().Printf("Received data with len %d", n)
		me.data.incomingChan <- buf[:n]
	}
}

func (me *Me) Outgoing() {
	defer me.Wg.Done()

	for {
		select {
		case <-me.Ctx.Done():
			return
		case data := <-me.data.outgoingChan:
			for name, peer := range me.peerset.Active {
				// Get or establish connection
				connI, ok := me.conns.Load(peer.P.Addr.String())
				if !ok {
					log.Default().Printf("Connecting to peer %s", name)
					conn, err := me.dialPeer(peer.P.Addr)
					if err != nil {
						log.Default().Printf("Failed to connect to peer %s: %v", name, err)
						continue
					}
					me.conns.Store(peer.P.Addr.String(), conn)
					connI = conn
				}
				conn := connI.(quic.Connection)

				// Open a new stream for sending data
				stream, err := conn.OpenStreamSync(context.Background())
				if err != nil {
					log.Default().Printf("Failed to open stream to %s: %v", name, err)
					continue
				}

				log.Default().Printf("Sending data to %s with len %d", name, len(data))
				_, err = stream.Write(data)
				if err != nil {
					log.Default().Printf("Error sending data to %s: %v", name, err)
				}
				stream.Close()
			}
		}
	}
}

func (me *Me) dialPeer(addr net.Addr) (quic.Connection, error) {
	return me.server.Dial(context.Background(), addr, me.tlsConfig, me.quicConfig)
}
