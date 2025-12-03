package peer

import (
	"log/slog"
	"net"
	"time"

	"github.com/vs-ude/btml/internal/model"
	"github.com/vs-ude/btml/internal/structs"
	"github.com/vs-ude/btml/internal/telemetry"
)

func Start(c *Config, m *model.Model, t *telemetry.Client) *Me {
	self := &structs.Peer{
		Name:        c.Name,
		Fingerprint: "abbabbaba",
	}
	me := NewMe(c, t, self)
	me.Setup()
	self.Addr = me.localAddr.(*net.UDPAddr)

	me.tracker = &Tracker{
		URL:        c.TrackerURL,
		UpdateFreq: c.UpdateFreq,
	}
	me.tracker.Setup(c, self)

	me.peerset = NewPeerSet(c.PeerSetSize, c.PeerSetArchiveAfter, me.telemetry)
	me.Wg.Add(1)
	go me.Listen()

	me.Wg.Add(1)
	go me.MaintenanceLoop()

	me.Wg.Add(1)
	go me.Outgoing()

	me.Wg.Add(1)
	go me.LaggingPeersLoop()

	return me
}

// WaitReady waits until we get at least one peer from the tracker.
func (me *Me) WaitReady() {
	for me.peerset.UnchokedLen() < 1 {
		time.Sleep(time.Second * 2)
	}
	slog.Info("Peer is ready")
}
