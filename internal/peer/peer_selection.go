package peer

import (
	"errors"
)

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
