package structs

import (
	"fmt"
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
	return fmt.Sprintf("%s (%s)", p.Name, p.Addr.String())
}

// Copy returns an independent deep copy of the Peer struct.
func (p *Peer) Copy() *Peer {
	return &Peer{
		Name:        p.Name,
		Addr:        p.Addr,
		Fingerprint: p.Fingerprint,
		LastSeen:    p.LastSeen,
	}
}
