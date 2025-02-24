package tracker

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/vs-ude/btml/internal/structs"
)

func (t *Tracker) list(w http.ResponseWriter, r *http.Request) {
	t.touchlist <- touch{r.Header.Get("peer-id"), time.Now()}
	data, _ := t.peers.Marshal()
	w.Write(data)
}

// join adds a peer to the peerlist. It will actually be added to the peerlist
// during the next maintenance cycle.
func (t *Tracker) join(w http.ResponseWriter, r *http.Request) {
	peer, err := getPeer(r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Default().Println("Error:", err.Error())
		return
	}
	peer.LastSeen = time.Now()
	t.newlist <- peer
	w.WriteHeader(http.StatusOK)
	log.Default().Printf("Added %s to the list of peers in the swarm\n", peer)
}

// leave removes a peer from the peerlist. It will actually be removed from the
// peerlist during the next maintenance cycle.
func (t *Tracker) leave(w http.ResponseWriter, r *http.Request) {
	peer, err := getPeer(r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Default().Println("Error:", err.Error())
		return
	}
	t.removelist <- peer.String()
	w.WriteHeader(http.StatusOK)
	log.Default().Printf("Removed %s from the list of peers in the swarm\n", peer)
}

// getPeer extracts the requesting peer from the request body.
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

// getPeerList returns a list of peers from the peerlist. If less than
// Tracker.conf.MaxReturnPeers are available, it will return all peers,
// otherwise it will return a randomized list of peers.
// The result may include peers that have already left the swarm.
func (t *Tracker) getPeerList() *structs.Peerlist {
	if t.peers.Len() <= t.conf.Tracker.MaxReturnPeers {
		t.processAddedPeers()
		return t.peers
	}

	i := t.conf.Tracker.MaxReturnPeers
	pl := &structs.Peerlist{}
	t.peers.Lock()
	defer t.peers.Unlock()
	// Pseudo-random iteration order is the default in Go
	for _, p := range t.peers.List {
		if i == 0 {
			break
		}
		// No need to lock as we are the only ones accessing _this_ peerlist.
		// We deep copy to prevent invalid accesses later during marshaling.
		// This may be too slow for large swarms with frequent requests.
		pl.List[p.String()] = p.Copy()
		i--
	}
	return pl
}
