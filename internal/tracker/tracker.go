package tracker

import (
	"log"
	"net/http"

	"github.com/BurntSushi/toml"
	"github.com/vs-ude/btfl/internal/structs"
)

type Tracker struct {
	addr  string
	peers *structs.Peerlist
	conf  *Config
}

func NewTracker(addr string, conf string) *Tracker {
	c := &Config{}
	if conf != "" {
		log.Default().Printf("Reading config from %s", conf)
		_, err := toml.DecodeFile(conf, c)
		if err != nil {
			log.Fatal(err)
		}
	}
	return &Tracker{
		addr: addr,
		conf: c,
	}
}

func (t *Tracker) Serve(done chan int) {
	log.Default().Printf("Running with config: %s", t.conf)
	t.peers = new(structs.Peerlist)
	t.peers.List = make(map[string]*structs.Peer)
	http.HandleFunc("/list", t.list)
	http.HandleFunc("/join", t.join)
	http.HandleFunc("/leave", t.leave)
	http.HandleFunc("/whoami", t.initPeer)
	log.Default().Printf("Tracker listening on http://%s", t.addr)
	log.Default().Println(http.ListenAndServe(t.addr, nil))
	done <- 1
}
