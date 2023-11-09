package structs

type Protocol int

const (
	UDP Protocol = iota
	TCP
	QUIC
)
