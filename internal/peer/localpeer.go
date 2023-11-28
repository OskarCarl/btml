package peer

import (
	"log"
	"net"
	"sync"
)

var localPeer = new(LocalPeer)

// LocalPeer is the peer we use
type LocalPeer struct {
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

func (l *LocalPeer) Setup() {
	server, err := net.ListenPacket("udp", "localhost:0")
	if err != nil {
		log.Default().Panicf("Error listening for packets: %v\n", err)
	}
	l.localAddr, _ = net.ResolveUDPAddr(server.LocalAddr().Network(), server.LocalAddr().String())
	l.server = server
	log.Default().Printf("Listening on %s", l.localAddr.String())
}

func (l *LocalPeer) Listen(wg *sync.WaitGroup, quit chan bool) {
	defer wg.Done()
	var (
		dchan = make(chan incomingPacket, 10)
		d     incomingPacket
	)

	go l.listen(dchan)
	select {
	case <-quit:
		l.server.Close()
		return
	case d = <-dchan:
		log.Default().Printf("Received packet from %s with len %d", d.from, len(d.data))
	}
}

func (l *LocalPeer) listen(dchan chan incomingPacket) {
	p := make([]byte, 1024)
	for {
		n, addr, err := l.server.ReadFrom(p)
		if err != nil {
			log.Default().Printf("Error when reading packet from %s: %v", addr, err)
		}
		d := make([]byte, n)
		copy(d, p[:n])
		packet := incomingPacket{
			from: addr,
			data: d,
		}
		dchan <- packet
	}
}

func (l *LocalPeer) Outgoing(dc chan []byte, wg *sync.WaitGroup, quit chan bool) {
	var (
		d   []byte
		err error
	)
	for {
		select {
		case <-quit:
			return
		case d = <-dc:
			n := 0
			for name, peer := range l.peerset.Active {
				log.Default().Printf("Sending data to %s: %v", name, d)
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
