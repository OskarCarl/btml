package structs

type Peer struct {
	Name,
	Addr string
	Proto       Protocol
	Fingerprint string
}

func (p *Peer) String() string {
	return p.Name
}
