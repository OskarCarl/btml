package structs

type Peer struct {
	Addr        string
	Proto       Protocol
	Fingerprint string
}

var NilPeer = &Peer{}

func (p *Peer) String() string {
	return p.Addr
}
