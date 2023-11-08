package main

import (
	"flag"

	"github.com/vs-ude/btfl/internal/tracker"
)

func main() {
	var listenAddr string
	flag.StringVar(&listenAddr, "ListenAddress", "localhost:8080", "The address the tracker listens on. Default: localhost:8080")

	tracker.Serve(listenAddr)
}
