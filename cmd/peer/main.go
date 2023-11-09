package main

import (
	"flag"

	"github.com/vs-ude/btfl/internal/logging"
	"github.com/vs-ude/btfl/internal/peer"
)

func main() {
	var trackerURL string
	flag.StringVar(&trackerURL, "trackerURL", "http://localhost:8080", "The URL of the tracker. Default is http://localhost:8080.")
	flag.Parse()

	logging.Logger.SetPrefix("[PEER]")
	logging.Logger.Use()

	peer.Start(trackerURL)
}
