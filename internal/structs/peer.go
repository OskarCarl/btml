package structs

import (
	"net"
	"time"
)

type Peer struct {
	Name        string
	Addr        *net.UDPAddr
	Fingerprint string
	LastSeen    time.Time
}

func (p *Peer) String() string {
	return p.Name
}
