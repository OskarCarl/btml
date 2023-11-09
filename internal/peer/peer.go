package peer

import (
	"log"
	"net"
	"time"

	"github.com/vs-ude/btfl/internal/structs"
)

func Start(trackerURL string) {
	server, err := net.ListenPacket("udp", "localhost:0")
	if err != nil {
		log.Default().Printf("Error listening for packets: %v\n", err)
		return
	}
	log.Default().Printf("Listening on %s", server.LocalAddr())

	tracker := &Tracker{
		URL: trackerURL,
		Identity: &structs.Peer{
			Addr:        server.LocalAddr().String(),
			Proto:       structs.UDP,
			Fingerprint: "abbabbaba",
		},
	}
	err = tracker.Join()
	if err != nil {
		log.Default().Printf("Error joining the tracker: %v\n", err)
		return
	}
	err = tracker.Update()
	if err != nil {
		log.Default().Printf("Error updating peers from the tracker: %v\n", err)
		tracker.Leave()
		return
	}
	time.Sleep(time.Second * 5)
	tracker.Leave()
}
