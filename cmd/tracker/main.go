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
	flag.StringVar(&listenAddr, "listen", "127.0.0.1:8080", "The address the tracker listens on. Default: 127.0.0.1:8080")
	flag.Parse()

	logging.Logger.SetPrefix("[TRACKER]")
	logging.Logger.Use()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	done := make(chan int, 1)
	t := &tracker.Tracker{
		Addr: listenAddr,
	}
	go t.Serve(done)
	go t.MaintenanceLoop()

	select {
	case <-sig:
		log.Default().Println("Tracker is terminating.")
		os.Exit(0)
	case exitCode := <-done:
		os.Exit(exitCode)
	}
}
