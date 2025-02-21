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
	m "github.com/vs-ude/btml/internal/model"
	"github.com/vs-ude/btml/internal/peer"
)

func main() {
	var trackerURL string
	var name string
	var python string
	var model string
	var datapath string
	var logpath string
	var autoconf bool
	flag.StringVar(&trackerURL, "tracker", "http://127.0.0.1:8080", "The URL of the tracker.")
	flag.StringVar(&name, "name", "", "Name of the peer. Default is a random int(0,100).")
	flag.StringVar(&python, "python", "python3", "Python runtime to use. Relative paths are based on the model path.")
	flag.StringVar(&model, "model", "model/", "Path where the main.py file is located.")
	flag.StringVar(&datapath, "datapath", "data/prepared/", "Base path for the training and testing data. Relative to the model path.")
	flag.StringVar(&logpath, "logpath", "logs/model.log", "Path for the python log file. Relative to the model path.")
	flag.BoolVar(&autoconf, "autoconf", false, "Automatically configure this peer using the provided tracker.")
	flag.Parse()

	c := &peer.Config{
		TrackerURL: trackerURL,
		ModelConf: &m.Config{
			PythonRuntime: python,
			ModelPath:     model,
			DataPath:      datapath,
			LogPath:       logpath,
		},
	}
	if autoconf {
		fmt.Printf("> Using peer autoconfiguration with tracker %s <\n", trackerURL)
		err := peer.Autoconf(c)
		if err != nil {
			fmt.Printf("Autoconfiguration failed: %v\n", err)
			os.Exit(1)
		}
	} else {
		if name == "" {
			i, _ := rand.Int(rand.Reader, big.NewInt(100))
			name = strconv.Itoa(int(i.Int64()))
		}
		c.Name = name
		c.UpdateFreq = time.Second * 10
		c.ModelConf.Name = name
	}

	logging.Logger.SetPrefix("[PEER " + c.Name + "]")
	logging.Logger.Use()

	m, err := m.NewSimpleModel(c.ModelConf)
	if err != nil {
		fmt.Printf("Failed to create model: %v\n", err)
		os.Exit(1)
	}

	peer.Start(c, m)
}
