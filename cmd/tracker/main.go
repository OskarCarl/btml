package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/vs-ude/btfl/internal/logging"
	"github.com/vs-ude/btfl/internal/tracker"
)

func main() {
	var listenAddr string
	flag.StringVar(&listenAddr, "ListenAddress", "localhost:8080", "The address the tracker listens on. Default: localhost:8080")
	flag.Parse()

	logging.Logger.SetPrefix("[TRACKER]")
	logging.Logger.Use()

	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)

	go tracker.Serve(listenAddr)

	<-done
	log.Default().Println("Tracker is terminating.")
	os.Exit(0)
}
