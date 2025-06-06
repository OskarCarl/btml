package tracker

import (
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/vs-ude/btml/internal/structs"
	"github.com/vs-ude/btml/internal/telemetry"
)

type touch struct {
	peer string
	ts   time.Time
}

type Tracker struct {
	addr           string
	peers          *structs.Peerlist
	blockedPeerIds struct {
		list []string
		sync.Mutex
	}
	conf      *Config
	telemetry struct {
		enabled bool
		ready   bool
	}
	newlist    chan *structs.Peer
	removelist chan string
	touchlist  chan touch
}

func NewTracker(addr string, conf string) *Tracker {
	c := &Config{}
	if conf != "" {
		slog.Debug("Reading config", "path", conf)
		_, err := toml.DecodeFile(conf, c)
		if err != nil {
			slog.Error("Failed to read config", "error", err)
			panic(err)
		}
	}
	// We assume that no more than 1000 peers will join/leave between two maintenance cycles
	t := &Tracker{
		addr: addr,
		conf: c,
		blockedPeerIds: struct {
			list []string
			sync.Mutex
		}{
			list: []string{},
		},
		telemetry: struct {
			enabled bool
			ready   bool
		}{
			enabled: false,
			ready:   false,
		},
		newlist:    make(chan *structs.Peer, 1000),
		removelist: make(chan string, 1000),
		touchlist:  make(chan touch, 10000),
	}
	return t
}

func (t *Tracker) EnableTelemetry() {
	t.telemetry.enabled = true
}

func (t *Tracker) SetupTelemetry(configPath string, done chan int) {
	err := telemetry.InitConf(t.conf.TelConf, t.conf.GrafanaConf, configPath)
	if err != nil {
		slog.Error("Telemtry is configured but could not be set up", "error", err)
		done <- 5
		return
	}
	t.telemetry.ready = true
}

func (t *Tracker) Serve(done chan int) {
	slog.Info("Starting tracker", "config", t.conf.String())
	t.peers = new(structs.Peerlist)
	t.peers.List = make(map[string]*structs.Peer)
	http.HandleFunc("/list", t.list)
	http.HandleFunc("/join", t.join)
	http.HandleFunc("/leave", t.leave)
	http.HandleFunc("/whoami", t.initPeer)
	slog.Info("Tracker listening", "addr", "http://"+t.addr)
	slog.Error("HTTP server terminated", "error", http.ListenAndServe(t.addr, nil))
	done <- 1
}
