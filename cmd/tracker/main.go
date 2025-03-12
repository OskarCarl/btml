package main

import (
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/vs-ude/btml/internal/logging"
	"github.com/vs-ude/btml/internal/tracker"
)

func main() {
	var listenAddr string
	var configPath string
	flag.StringVar(&listenAddr, "listen", ":8080", "The address the tracker listens on.")
	flag.StringVar(&configPath, "config", "config.toml", "The path to the configuration file.")
	flag.Parse()

	logging.FromEnv()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	done := make(chan int, 1)
	t := tracker.NewTracker(listenAddr, configPath)
	go t.Serve(done)
	go t.MaintenanceLoop()

	select {
	case <-sig:
		slog.Info("Tracker is terminating")
		os.Exit(0)
	case exitCode := <-done:
		os.Exit(exitCode)
	}
}
