package peer

import (
	"errors"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/vs-ude/btfl/internal/structs"
)

func (kp *KnownPeer) Connect() (err error) {
	switch kp.P.Proto {
	case structs.UDP:
		kp.C = new(udpConn)
		err = kp.C.Connect(kp.P.Addr)
	default:
		return errors.ErrUnsupported
	}
	return err
}

type PeerConnection interface {
	Connect(addr string) error
	Close() error
	Send(b []byte) error
}

type udpConn struct {
	conn *net.UDPConn
}

func (c *udpConn) Connect(addr string) error {
	conn, err := net.DialUDP("udp", udpParse(localPeer.localAddr.String()), udpParse(addr))
	if err != nil {
		return err
	}
	c.conn = conn
	return nil
}

func udpParse(s string) *net.UDPAddr {
	tmp := strings.Split(s, ":")
	ip := net.ParseIP(tmp[0])
	port, _ := strconv.Atoi(tmp[1])
	return &net.UDPAddr{
		IP:   ip,
		Port: port,
	}
}

func (c *udpConn) Close() error {
	return c.conn.Close()
}

func (c *udpConn) Send(b []byte) error {
	// TODO: need to encode lots of metadata. Use gob?
	c.conn.SetWriteDeadline(time.Now().Add(time.Second * 10))
	_, err := c.conn.Write(b)
	return err
}
