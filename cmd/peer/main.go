package main

import (
	"crypto/rand"
	"flag"
	"fmt"
	"math/big"

	"github.com/vs-ude/btfl/internal/logging"
	"github.com/vs-ude/btfl/internal/peer"
)

func main() {
	var trackerURL string
	var name string
	flag.StringVar(&trackerURL, "trackerURL", "http://127.0.0.1:8080", "The URL of the tracker. Default is http://127.0.0.1:8080.")
	flag.StringVar(&name, "name", "", "Name of the peer. Default is a random int.")
	flag.Parse()

	if name == "" {
		i, _ := rand.Int(rand.Reader, big.NewInt(10000))
		name = fmt.Sprintf("%d", i)
	}

	logging.Logger.SetPrefix("[PEER " + name + "]")
	logging.Logger.Use()

	c := &peer.Config{
		Name:       name,
		TrackerURL: trackerURL,
	}
	peer.Start(c)
}
