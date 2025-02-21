package peer

import (
	"crypto/rand"
	"log"
	"math/big"
	"sync"
	"time"

	"github.com/vs-ude/btml/internal/model"
	"github.com/vs-ude/btml/internal/structs"
)

func Start(c *Config, m model.Model) {
	localPeer.config = c
	localPeer.Setup()
	self := &structs.Peer{
		Name:        c.Name,
		Fingerprint: "abbabbaba",
	}
	self.Addr = localPeer.localAddr

	tracker := &Tracker{
		URL:        c.TrackerURL,
		UpdateFreq: c.UpdateFreq,
	}
	tracker.Setup(c, self)
	defer tracker.Leave()

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
	num := 0
	for _, p := range localPeer.tracker.Peers.List {
		if num > 4 {
			break
		}
		localPeer.peerset.Add(p)
		num++
	}

	go m.Train()
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

// periodicUpdate periodically updates the peer list from the tracker.
// This has the side effect of pinging the tracker so it knows we are alive.
func periodicUpdate(t *Tracker, wg *sync.WaitGroup, done chan struct{}) {
	defer wg.Done()
	timer := time.NewTimer(time.Second)
	for {
		select {
		case <-timer.C:
			err := t.Update()
			if err != nil {
				log.Default().Printf("Error updating peers from the tracker: %v\n", err)
				return
			}
			timer.Reset(t.UpdateFreq)
		case <-done:
			return
		}
	}
}
