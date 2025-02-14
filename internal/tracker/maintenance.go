package tracker

import (
	"log"
	"time"
)

func (t *Tracker) MaintenanceLoop() {
	log.Default().Println("Starting maintenance loop")
	for {
		time.Sleep(time.Second * 30)
		log.Default().Println("Running periodic maintenance")
		t.cleanPeers()
	}
}

// cleanPeers removes all peers from the peerlist which have not sent an update
// for the last t.conf.PeerTimeout duration.
func (t *Tracker) cleanPeers() {
	for _, p := range t.peers.List {
		if p.LastSeen.Before(time.Now().Add(-t.conf.PeerTimeout)) {
			log.Default().Printf("Removing %s from peerset due to inactivity", p)
			t.peers.Remove(p)
		}
	}
}
