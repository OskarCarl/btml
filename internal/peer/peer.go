package peer

import (
	"fmt"
	"net"
	"time"
)

func Start(trackerURL string) {
	server, err := net.ListenPacket("udp", "localhost:0")
	if err != nil {
		fmt.Printf("Error listening for packets: %v\n", err)
		return
	}

	tracker := &Tracker{
		URL:        trackerURL,
		ListenAddr: server.LocalAddr().String(),
	}
	err = tracker.Join()
	if err != nil {
		fmt.Printf("Error joining the tracker: %v\n", err)
		return
	}
	err = tracker.Update()
	if err != nil {
		fmt.Printf("Error updating peers from the tracker: %v\n", err)
		tracker.Leave()
		return
	}
	time.Sleep(time.Second * 5)
	tracker.Leave()
}
