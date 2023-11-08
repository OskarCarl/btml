package main

import (
	"flag"

	client "github.com/vs-ude/btfl/internal/peer"
)

func main() {
	var trackerURL string
	flag.StringVar(&trackerURL, "trackerURL", "http://localhost:8080", "The URL of the tracker. Default is http://localhost:8080.")
	client.Start(trackerURL)
}
