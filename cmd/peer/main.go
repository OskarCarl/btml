package main

import (
	"crypto/rand"
	"flag"
	"fmt"
	"math/big"
	"os"
	"strconv"

	"github.com/vs-ude/btfl/internal/logging"
	"github.com/vs-ude/btfl/internal/peer"
)

func main() {
	var trackerURL string
	var name string
	var basepath string
	var logpath string
	var autoconf bool
	flag.StringVar(&trackerURL, "trackerURL", "http://127.0.0.1:8080", "The URL of the tracker. Default is http://127.0.0.1:8080.")
	flag.StringVar(&name, "name", "", "Name of the peer. Default is a random int.")
	flag.StringVar(&basepath, "basepath", "", "Base path for the peer. Default is the current working directory.")
	flag.StringVar(&logpath, "logpath", "", "Base path for the peer logs. Default is the current working directory.")
	flag.BoolVar(&autoconf, "autoconf", false, "Automatically configure this peer using the provided tracker.")
	flag.Parse()

	if basepath == "" {
		basepath, _ = os.Getwd()
	}

	if logpath == "" {
		logpath = basepath
	}

	c := &peer.Config{
		TrackerURL: trackerURL,
		Basepath:   basepath,
		Logpath:    logpath,
	}
	if autoconf {
		fmt.Printf("Using peer autoconfiguration with tracker %s\n", trackerURL)
		err := peer.Autoconf(c)
		if err != nil {
			fmt.Printf("Autoconfiguration failed: %v\n", err)
			os.Exit(1)
		}
	} else {
		if name == "" {
			i, _ := rand.Int(rand.Reader, big.NewInt(10000))
			name = strconv.Itoa(int(i.Int64()))
		}
		c.Name = name
	}

	logging.Logger.SetPrefix("[PEER " + c.Name + "]")
	logging.Logger.Use()

	peer.Start(c)
}
