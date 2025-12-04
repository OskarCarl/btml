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
	selection := make(map[string]*KnownPeer, me.config.PeerSetSize)
	// Select new peers
	for n, _ := range me.tracker.Peers.List {
		selection[n] = me.peerset.known[n]
		i++
		if i == me.peerset.softMaxSize {
			break
		}
	}

	// Choke previous peers which are not selected
	for n, p := range me.peerset.unchoked {
		if _, ok := selection[n]; !ok {
			p.choke()
		}
	}

	me.peerset.unchoked = selection
	return nil
}

type DefaultBittorrentPeerSelectionStrategy struct {
}

func (btps *DefaultBittorrentPeerSelectionStrategy) Select(me *Me) error {
	if len(me.tracker.Peers.List) == 0 {
		return errors.New("No peers available")
	}
	// i := 0
	selection := make(map[string]*KnownPeer, me.config.PeerSetSize)
	// Select new peers based on score
	// sort.Slice(me.tracker.Peers.List, func(i, j int) bool {
	// 	peer_i, ok_i := me.peerset.Known[string(i)]
	// 	peer_j, ok_j := me.peerset.Known[string(j)]
	// 	switch {
	// 	case !ok_i || !ok_j:
	// 		return false
	// 	case !ok_i:
	// 		return true
	// 	case !ok_j:
	// 		return false
	// 	default:
	// 		return peer_i.score < peer_j.score
	// 	}
	// })
	// for n, p := range me.tracker.Peers.List {
	// 	if p.Score > 0 {
	// 		selection[n] = me.peerset.Known[n]
	// 		i++
	// 		if i == me.peerset.SoftMaxSize {
	// 			break
	// 		}
	// 	}
	// }

	// Choke previous peers which are not selected
	for n, p := range me.peerset.unchoked {
		if _, ok := selection[n]; !ok {
			p.choke()
		}
	}

	me.peerset.unchoked = selection
	return nil
}
