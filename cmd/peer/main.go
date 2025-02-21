package main

import (
	"crypto/rand"
	"flag"
	"fmt"
	"math/big"
	"os"
	"strconv"
	"time"

	"github.com/vs-ude/btml/internal/logging"
	"github.com/vs-ude/btml/internal/peer"
)

func main() {
	var trackerURL string
	var name string
	var model string
	var datapath string
	var logpath string
	var autoconf bool
	flag.StringVar(&trackerURL, "tracker", "http://127.0.0.1:8080", "The URL of the tracker. Default is http://127.0.0.1:8080.")
	flag.StringVar(&name, "name", "", "Name of the peer. Default is a random int.")
	flag.StringVar(&model, "model", "model/", "Path where the main.py file is located. Default is model/.")
	flag.StringVar(&datapath, "datapath", "model/data/", "Base path for the training and testing data. Default is model/data/.")
	flag.StringVar(&logpath, "logpath", "model/logs/model.log", "Path for the python log file. Default is model/logs/model.log.")
	flag.BoolVar(&autoconf, "autoconf", false, "Automatically configure this peer using the provided tracker.")
	flag.Parse()

	c := &peer.Config{
		TrackerURL: trackerURL,
		Datapath:   datapath,
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
		c.UpdateFreq = time.Second * 10
	}

	logging.Logger.SetPrefix("[PEER " + c.Name + "]")
	logging.Logger.Use()

	peer.Start(c)
}
