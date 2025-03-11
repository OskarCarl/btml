package main

import (
	"crypto/rand"
	"flag"
	"fmt"
	"log"
	"math/big"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/vs-ude/btml/internal/logging"
	"github.com/vs-ude/btml/internal/model"
	"github.com/vs-ude/btml/internal/peer"
	"github.com/vs-ude/btml/internal/play"
)

func main() {
	var trackerURL string
	var name string
	var python string
	var modelPath string
	var datapath string
	var logpath string
	var autoconf bool
	flag.StringVar(&trackerURL, "tracker", "http://127.0.0.1:8080", "The URL of the tracker.")
	flag.StringVar(&name, "name", "", "Name of the peer. Default is a random int(0,100).")
	flag.StringVar(&python, "python", "python3", "Python runtime to use. Relative paths are based on the model path.")
	flag.StringVar(&modelPath, "model", "model/", "Path where the main.py file is located.")
	flag.StringVar(&datapath, "datapath", "data/prepared/", "Base path for the training and testing data. Relative to the model path.")
	flag.StringVar(&logpath, "logpath", "logs/model.log", "Path for the python log file. Relative to the model path.")
	flag.BoolVar(&autoconf, "autoconf", false, "Automatically configure this peer using the provided tracker.")
	flag.Parse()

	c := &peer.Config{
		TrackerURL: trackerURL,
		ModelConf: &model.Config{
			PythonRuntime: python,
			ModelPath:     modelPath,
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
		c.Addr = "127.0.0.1"
		c.UpdateFreq = time.Second * 10
		c.ModelConf.Name = name
	}

	logging.Logger.SetPrefix("[PEER " + c.Name + "]")
	logging.Logger.Use()

	os.Exit(run(c))
}

func run(c *peer.Config) int {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	m, err := model.NewModel(c.ModelConf)
	if err != nil {
		fmt.Printf("Failed to create model: %v\n", err)
		os.Exit(1)
	}
	err = m.Start()
	if err != nil {
		fmt.Printf("Failed to start model: %v\n", err)
		os.Exit(1)
	}
	defer m.Shutdown()

	me := peer.Start(c, m)
	defer me.Shutdown()

	strategy := model.NewNaiveStrategy(m)

	ch, _ := me.ListenForWeights()
	strategy.Start(ch)
	m.SetCallback(func(weights *model.Weights) {
		i, _ := rand.Int(rand.Reader, big.NewInt(100))
		if i.Cmp(big.NewInt(50)) < 0 {
			return
		}
		me.DistributeWeights(weights)
	})

	go localPlay(m, me)

	select {
	case <-sig:
		log.Default().Println("Peer is terminating")
		return 0
	case <-me.Ctx.Done():
		return 2
	}
}

func localPlay(m *model.Model, peer *peer.Me) {
	log.Default().Printf("Starting local play")
	p := play.NewPlay(peer, m)
	p.AddStep(&play.Train{})
	p.AddStep(&play.Eval{})
	p.AddStep(&play.Wait{T: time.Second * 10})
	p.AddStep(&play.Train{})
	p.AddStep(&play.Wait{T: time.Second * 5})
	p.AddStep(&play.Train{})
	p.AddStep(&play.Wait{T: time.Second * 4})
	p.AddStep(&play.Train{})
	p.AddStep(&play.Wait{T: time.Second * 3})

	peer.WaitReady()
	p.Run()
}
