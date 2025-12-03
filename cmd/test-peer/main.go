package main

import (
	"crypto/rand"
	"flag"
	"fmt"
	"log/slog"
	"math/big"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/vs-ude/btml/internal/logging"
	"github.com/vs-ude/btml/internal/model"
	"github.com/vs-ude/btml/internal/peer"
	"github.com/vs-ude/btml/internal/structs"
)

func main() {
	var name string
	var port int
	var peers string
	flag.StringVar(&name, "name", "", "Name of the peer. Default is a random int(0,100).")
	flag.IntVar(&port, "port", 0, "Port to listen on. Default is a random port.")
	flag.StringVar(&peers, "peers", "", "Comma-separated list of peers to connect to.")
	flag.Parse()

	logging.FromEnv()

	c := &peer.Config{
		Addr:       fmt.Sprintf(":%d", port),
		TrackerURL: "",
		ModelConf:  nil,
		UpdateFreq: -1,
	}
	if name == "" {
		i, _ := rand.Int(rand.Reader, big.NewInt(100))
		name = strconv.Itoa(int(i.Int64()))
	}
	c.Name = name
	c.ModelConf = &model.Config{
		Name: name,
	}
	logging.SetID(c.Name)

	ps := manualPeerSet(peers)

	os.Exit(run(c, ps))
}

func run(c *peer.Config, ps *peer.PeerSet) int {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	me := peer.Start(c, nil, nil)
	defer me.Shutdown()

	var strategy model.ApplyStrategy = &DummyStrategy{}
	ch, _ := me.ListenForWeights()
	strategy.Start(ch)
	me.ManualPeerSet(ps)

	go dummySend(me)

	select {
	case <-sig:
		slog.Info("Peer is terminating")
		return 0
	case <-me.Ctx.Done():
		return 2
	}
}

func dummySend(p *peer.Me) {
	for i := range 100 {
		w := &model.Weights{}
		w.SetAge(i)
		p.Send(w)
		time.Sleep(time.Second * 10)
	}

}

func manualPeerSet(list string) *peer.PeerSet {
	l := strings.Split(list, ",")
	ps := peer.NewPeerSet(len(list), time.Hour, nil)
	for _, p := range l {
		addr, err := net.ResolveUDPAddr("udp", p)
		if err != nil {
			slog.Error("Failed to resolve address", "err", err)
			continue
		}
		ps.Add(&structs.Peer{
			Name: p,
			Addr: addr,
		})
	}
	return ps
}

type DummyStrategy struct{}

func (d *DummyStrategy) SetModel(m *model.Model) {}
func (d *DummyStrategy) Start(wc <-chan *model.Weights) error {
	go func() {
		for w := range wc {
			slog.Info("Received message", "age", w.GetAge())
		}
	}()
	return nil
}
