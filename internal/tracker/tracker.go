package tracker

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/vs-ude/btfl/internal/structs"
)

type tracker struct {
	peers structs.Peerlist
}

func Serve(listenAddr string) {
	t := &tracker{peers: *structs.NewPeerList()}
	http.HandleFunc("/list", t.list)
	http.HandleFunc("/join", t.join)
	http.HandleFunc("/leave", t.leave)
	log.Default().Println("Starting")
	log.Default().Println(http.ListenAndServe(listenAddr, nil))
}

func (t *tracker) list(w http.ResponseWriter, r *http.Request) {
	data, _ := t.peers.Marshal()
	w.Write(data)
}

func (t *tracker) join(w http.ResponseWriter, r *http.Request) {
	peer, err := getPeer(r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Default().Println("Error:", err.Error())
		return
	}
	t.peers.Add(peer)
	w.WriteHeader(http.StatusOK)
	log.Default().Printf("Added %s to the list of peers in the swarm\n", peer)
}

func (t *tracker) leave(w http.ResponseWriter, r *http.Request) {
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
		return structs.NilPeer, fmt.Errorf("unable to read request body data from %s: %s\n%w", r.RemoteAddr, r.RequestURI, err)
	}
	var peer structs.Peer
	err = json.Unmarshal(body[0:n], &peer)
	if err != nil {
		return structs.NilPeer, fmt.Errorf("unable to parse request body data from %s: %s\n%w", r.RemoteAddr, r.RequestURI, err)
	}
	return &peer, nil
}
