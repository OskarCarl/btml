package peer

import (
	"time"
)

func (me *Me) MaintenanceLoop() {
	defer me.Wg.Done() // This is for the Add call before the function!
	me.Wg.Add(1)       // This is for periodicUpdate
	go me.tracker.periodicUpdate(&me.Wg, me.Ctx)

	timer := time.NewTimer(time.Second)
	wait := time.Duration(time.Second * 30)
	for {
		select {
		case <-me.Ctx.Done():
			return
		case <-timer.C:
			me.UpdatePeerset()
			if len(me.tracker.Peers.List) > 0 {
				me.pss.Select(me)
				me.sendTelemetry()
				timer.Reset(wait)
			} else {
				timer.Reset(time.Duration(time.Second * 5))
			}
		}
	}
}

func (me *Me) UpdatePeerset() {
	me.tracker.Lock()
	defer me.tracker.Unlock()
	for _, p := range me.tracker.Peers.List {
		me.peerset.Add(p)
	}
}

func (me *Me) sendTelemetry() {
	me.telemetry.RecordActivePeers(me.peerset.ActiveToString())
}
