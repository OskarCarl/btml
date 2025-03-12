package peer

import (
	"log"
	"net"
	"time"

	"github.com/vs-ude/btml/internal/model"
	"github.com/vs-ude/btml/internal/structs"
)

func Start(c *Config, m *model.Model) *Me {
	me := NewMe(c)
	me.Setup()
	self := &structs.Peer{
		Name:        c.Name,
		Fingerprint: "abbabbaba",
	}
	self.Addr = me.localAddr.(*net.UDPAddr)

	me.tracker = &Tracker{
		URL:        c.TrackerURL,
		UpdateFreq: c.UpdateFreq,
	}
	me.tracker.Setup(c, self)

	me.peerset = NewPeerSet(c.PeersetSize)
	me.Wg.Add(1)
	go me.Listen()

	me.Wg.Add(1)
	go me.MaintenanceLoop()

	me.Wg.Add(1)
	go me.Outgoing()

	return me
}

// WaitReady waits until we get at least one peer from the tracker. It adds up
// to 5 peers to the peer set.
func (me *Me) WaitReady() {
	for len(me.tracker.Peers.List) < 1 {
		time.Sleep(time.Second * 2)
	}
	me.UpdatePeerset()
	me.pss.Select(me)
	log.Default().Printf("Ready with %d peers", len(me.peerset.Active))
}
