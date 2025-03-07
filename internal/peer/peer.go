package peer

import (
	"net"
	"time"

	"github.com/vs-ude/btml/internal/model"
	"github.com/vs-ude/btml/internal/structs"
)

func Start(c *Config, m model.Model) *Me {
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

	me.peerset = NewPeerSet()
	me.Wg.Add(1)
	go me.Listen()

	me.Wg.Add(1)
	go me.tracker.periodicUpdate(&me.Wg, me.Ctx)

	me.Wg.Add(1)
	go me.Outgoing()

	return me
}

func (me *Me) Send(w model.Weights) {
	me.data.outgoingChan <- w
}

func (me *Me) WaitReady() {
	for len(me.tracker.Peers.List) < 1 {
		time.Sleep(time.Second * 2)
	}
	num := 0
	for _, p := range me.tracker.Peers.List {
		if num > 4 {
			break
		}
		me.peerset.Add(p)
		num++
	}
}
