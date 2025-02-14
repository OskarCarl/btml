package tracker

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/vs-ude/btfl/internal/structs"
)

type Tracker struct {
	Addr  string
	peers *structs.Peerlist
	conf  *Config
}

func (t *Tracker) Serve(done chan int) {
	t.peers = new(structs.Peerlist)
	t.peers.List = make(map[string]*structs.Peer)
	t.conf = &Config{
		PeerTimeout: time.Second * 30,
	}
	http.HandleFunc("/list", t.list)
	http.HandleFunc("/join", t.join)
	http.HandleFunc("/leave", t.leave)
	http.HandleFunc("/whoami", t.initPeer)
	log.Default().Printf("Starting on http://%s", t.Addr)
	log.Default().Println(http.ListenAndServe(t.Addr, nil))
	done <- 1
}

func (t *Tracker) list(w http.ResponseWriter, r *http.Request) {
	t.peers.Touch(r.Header.Get("peer-id"))
	data, _ := t.peers.Marshal()
	w.Write(data)
}

func (t *Tracker) join(w http.ResponseWriter, r *http.Request) {
	peer, err := getPeer(r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Default().Println("Error:", err.Error())
		return
	}
	peer.LastSeen = time.Now()
	t.peers.Add(peer)
	w.WriteHeader(http.StatusOK)
	log.Default().Printf("Added %s to the list of peers in the swarm\n", peer)
}

func (t *Tracker) leave(w http.ResponseWriter, r *http.Request) {
	peer, err := getPeer(r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Default().Println("Error:", err.Error())
		return
	}
	t.peers.Remove(peer)
	w.WriteHeader(http.StatusOK)
	log.Default().Printf("Removed %s from the list of peers in the swarm\n", peer)
}

func getPeer(r *http.Request) (*structs.Peer, error) {
	body := make([]byte, 1024)
	n, err := r.Body.Read(body)
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, fmt.Errorf("unable to read request body data from %s: %s\n%w", r.RemoteAddr, r.RequestURI, err)
	}
	var peer structs.Peer
	err = json.Unmarshal(body[0:n], &peer)
	if err != nil {
		return nil, fmt.Errorf("unable to parse request body data from %s: %s\n%w", r.RemoteAddr, r.RequestURI, err)
	}
	return &peer, nil
}
