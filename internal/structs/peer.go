package structs

import "net"

type Peer struct {
	Name        string
	Addr        *net.UDPAddr
	Fingerprint string
}

func (p *Peer) String() string {
	return p.Name
}
