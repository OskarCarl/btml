package peer

import (
	"log"
	"sync"
	"time"
)

const TRACKER_REFRESH = time.Second * 10

func Start(c *Config) {
	localPeer.config = c
	localPeer.Setup()

	tracker := new(Tracker)
	tracker.Setup(c)

	quit := make(chan bool, 1)
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go periodicUpdate(tracker, wg, quit)

	localPeer.tracker = tracker
	localPeer.peerset = NewPeerSet()
	wg.Add(1)
	go localPeer.Listen(wg, quit)

	ping()
	close(quit)
	wg.Wait()
}

func ping() {
	for len(localPeer.tracker.Peers.List) < 1 {
		time.Sleep(time.Second * 2)
	}
	for _, p := range localPeer.tracker.Peers.List {
		localPeer.peerset.Add(&p) // FIXME: This something is broken here, we get a nil pointer?
		kp := localPeer.peerset.Active["0"]
		kp.Connect()
		kp.C.Send([]byte{0xff, 0xaf})
	}
}

func periodicUpdate(t *Tracker, wg *sync.WaitGroup, done chan bool) {
	defer wg.Done()
	timer := time.NewTimer(time.Second)
	for {
		select {
		case <-timer.C:
			err := t.Update()
			if err != nil {
				log.Default().Printf("Error updating peers from the tracker: %v\n", err)
				t.Leave()
				return
			}
			timer.Reset(TRACKER_REFRESH)
		case <-done:
			t.Leave()
			return
		}
	}
}
