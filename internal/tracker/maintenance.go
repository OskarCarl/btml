package tracker

import (
	"log"
	"time"
)

func (t *Tracker) MaintenanceLoop() {
	log.Default().Println("Starting maintenance loop")
	for {
		time.Sleep(t.conf.Tracker.MaintainInterval)
		log.Default().Println("Running periodic maintenance")
		t.processAddedPeers()
		t.processTouches()
		t.processRemovedPeers()
		t.cleanPeers()
	}
}

// processTouches updates the pending last seen touches. It assumes the peer
// exists and should be run after the peer has been added to the peerlist.
func (t *Tracker) processTouches() {
	// We lock and unlock manually to make this more efficient
	t.peers.Lock()
	defer t.peers.Unlock()
	for {
		select {
		// We can assume the touches to be in the order they occurred
		case individualTouch := <-t.touchlist:
			if p, ok := t.peers.List[individualTouch.peer]; ok {
				p.LastSeen = individualTouch.ts
			}
		default:
			return
		}
	}
}

// cleanPeers removes all peers from the peerlist which have not sent an update
// for the last Tracker.conf.PeerTimeout duration.
func (t *Tracker) cleanPeers() {
	for _, p := range t.peers.List {
		if p.LastSeen.Before(time.Now().Add(-t.conf.Tracker.PeerTimeout)) {
			log.Default().Printf("Removing %s from peerset due to inactivity", p)
			t.peers.Remove(p)
		}
	}
}

// processAddedPeers adds all currently waiting peers to the peerlist.
func (t *Tracker) processAddedPeers() {
	// We lock and unlock manually to make this more efficient
	t.peers.Lock()
	defer t.peers.Unlock()
	for {
		select {
		case peer := <-t.newlist:
			t.peers.List[peer.Name] = peer
		default:
			return
		}
	}
}

// processRemovedPeers removes all peers from the peerlist that have sent an explicit
// leave message.
func (t *Tracker) processRemovedPeers() {
	// We lock and unlock manually to make this more efficient
	t.peers.Lock()
	defer t.peers.Unlock()
	for {
		select {
		case peerID := <-t.removelist:
			delete(t.peers.List, peerID)
		default:
			return
		}
	}
}
