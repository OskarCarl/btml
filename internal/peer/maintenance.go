package peer

import (
	"errors"
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
			me.pss.Select(me)
			timer.Reset(wait)
		}
	}
}

type PeerSelectionStrategy interface {
	Select(*Me) error
}

type RandomPeerSelectionStrategy struct {
}

func (rps *RandomPeerSelectionStrategy) Select(me *Me) error {
	if len(me.tracker.Peers.List) == 0 {
		return errors.New("No peers available")
	}
	i := 0
	selection := make(map[string]*KnownPeer, me.config.PeersetSize)
	// Select new peers
	for n, _ := range me.tracker.Peers.List {
		selection[n] = me.peerset.Known[n]
		i++
		if i == me.config.PeersetSize {
			break
		}
	}

	// Choke previous peers which are not selected
	for n, p := range me.peerset.Active {
		if _, ok := selection[n]; !ok {
			p.choke()
		}
	}

	me.peerset.Active = selection
	return nil
}
