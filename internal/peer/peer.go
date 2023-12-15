package peer

import (
	"crypto/rand"
	"log"
	"math/big"
	"sync"
	"time"
)

const TRACKER_REFRESH = time.Second * 10

func Start(c *Config) {
	localPeer.config = c
	localPeer.Setup()

	tracker := new(Tracker)
	tracker.Setup(c)

	quit := make(chan struct{})
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go periodicUpdate(tracker, wg, quit)

	localPeer.tracker = tracker
	localPeer.peerset = NewPeerSet()
	wg.Add(1)
	go localPeer.Listen(wg, quit)
	wg.Add(1)
	outgoingDataChan := make(chan []byte, 20)
	go localPeer.Outgoing(outgoingDataChan, wg, quit)

	for len(localPeer.tracker.Peers.List) < 1 {
		time.Sleep(time.Second * 2)
	}
	for _, p := range localPeer.tracker.Peers.List {
		localPeer.peerset.Add(&p)
	}

	go ping(outgoingDataChan)

	time.Sleep(time.Second * 60)
	close(quit)
	wg.Wait()
}

func ping(dc chan []byte) {
	for {
		wait, _ := rand.Int(rand.Reader, big.NewInt(5))
		time.Sleep(time.Second * time.Duration(big.NewInt(0).Add(wait, big.NewInt(2)).Int64()))
		dc <- []byte{0xff, 0xaf}
	}
}

func periodicUpdate(t *Tracker, wg *sync.WaitGroup, done chan struct{}) {
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
