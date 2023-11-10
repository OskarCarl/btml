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
	localAddr net.Addr
	server    net.PacketConn
	tracker   *Tracker
	peerset   *PeerSet
}

func (l *LocalPeer) Setup() {
	server, err := net.ListenPacket("udp", "localhost:0")
	if err != nil {
		log.Default().Panicf("Error listening for packets: %v\n", err)
	}
	l.localAddr = server.LocalAddr()
	l.server = server
	log.Default().Printf("Listening on %s", l.localAddr.String())
}

func (l *LocalPeer) Listen(wg *sync.WaitGroup, quit chan bool) {
	defer wg.Done()
	var (
		dchan = make(chan dataPacket, 10)
		d     dataPacket
	)

	go l.listen(dchan)
	select {
	case d = <-dchan:
		log.Default().Printf("Received packet from %s with len %d", d.from, len(d.data))
	case <-quit:
		l.server.Close()
		return
	}
}

func (l *LocalPeer) listen(dchan chan dataPacket) {
	p := make([]byte, 1024)
	for {
		n, addr, err := l.server.ReadFrom(p)
		if err != nil {
			log.Default().Printf("Error when reading packet from %s", addr)
		}
		d := make([]byte, n)
		copy(d, p[:n])
		packet := dataPacket{
			from: addr,
			data: d,
		}
		dchan <- packet
	}
}
