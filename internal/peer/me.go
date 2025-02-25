package peer

import (
	"errors"
	"log"
	"net"
	"sync"
	"time"
)

var me = new(Me)

// Me is the peer we use
type Me struct {
	config    *Config
	localAddr *net.UDPAddr
	server    net.PacketConn
	tracker   *Tracker
	peerset   *PeerSet
}

type incomingPacket struct {
	data []byte
	from net.Addr
}

func (l *Me) Setup() {
	server, err := net.ListenPacket("udp", ":0")
	if err != nil {
		log.Default().Panicf("Error listening for packets: %v\n", err)
	}
	l.localAddr, _ = net.ResolveUDPAddr(server.LocalAddr().Network(), server.LocalAddr().String())
	l.server = server
	log.Default().Printf("Listening on %s", l.localAddr.String())
}

func (l *Me) Listen(wg *sync.WaitGroup, quit chan struct{}) {
	defer func() {
		l.server.Close()
		wg.Done()
	}()
	var (
		dchan = make(chan incomingPacket, 10)
		d     incomingPacket
	)

	go l.listen(dchan)
	for {
		select {
		case <-quit:
			log.Default().Print("Stopping the listener...")
			return
		case d = <-dchan:
			log.Default().Printf("Received packet from %s with len %d", d.from, len(d.data))
		}
	}
}

func (l *Me) listen(dchan chan incomingPacket) {
	p := make([]byte, 1024)
	for {
		n, addr, err := l.server.ReadFrom(p)
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return
			}
			log.Default().Printf("Error when reading packet from %s: %v", addr, err)
			log.Default().Print("Continuing...")
		}
		d := make([]byte, n)
		log.Default().Print("Got packet")
		copy(d, p[:n])
		packet := incomingPacket{
			from: addr,
			data: d,
		}
		dchan <- packet
	}
}

func (l *Me) Outgoing(dc chan []byte, wg *sync.WaitGroup, quit chan struct{}) {
	defer wg.Done()
	var (
		d   []byte
		err error
	)
	for {
		select {
		case <-quit:
			return
		case d = <-dc:
			time.Sleep(time.Second * 20)
			for name, peer := range l.peerset.Active {
				n := 0
				log.Default().Printf("Sending data to %s with len %d", name, len(d))
				for n < len(d) {
					n, err = l.server.WriteTo(d, peer.P.Addr)
					// We might get an error here during shutdown when the server is closed by the listener.
					if err != nil {
						log.Default().Printf("Encountered an error when writing %d bytes to %s:\n%v", n, name, err)
						break
					}
				}
			}
		}
	}
}
