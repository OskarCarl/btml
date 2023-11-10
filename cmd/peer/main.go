package main

import (
	"flag"

	"github.com/vs-ude/btfl/internal/logging"
	"github.com/vs-ude/btfl/internal/peer"
)

func main() {
	var trackerURL string
	var name string
	flag.StringVar(&trackerURL, "trackerURL", "http://localhost:8080", "The URL of the tracker. Default is http://localhost:8080.")
	flag.StringVar(&name, "name", "peer", "Name of the peer. Default is 'peer'.")
	flag.Parse()

	logging.Logger.SetPrefix("[PEER " + name + "]")
	logging.Logger.Use()

	c := &peer.Config{
		Name:       name,
		TrackerURL: trackerURL,
	}
	peer.Start(c)
}
