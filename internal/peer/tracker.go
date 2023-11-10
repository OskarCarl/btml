package peer

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/vs-ude/btfl/internal/structs"
)

type Tracker struct {
	URL      string
	Peers    *structs.Peerlist
	Identity *structs.Peer
}

func (t *Tracker) Setup(c *Config) {
	t.URL = c.TrackerURL
	t.Identity = &structs.Peer{
		Name:        c.Name,
		Addr:        localPeer.localAddr.String(),
		Proto:       structs.UDP,
		Fingerprint: "abbabbaba",
	}
	t.Peers = structs.NewPeerList()

	err := t.Join()
	if err != nil {
		log.Default().Printf("Error joining the tracker: %v\n", err)
		return
	}
}

func (t *Tracker) Update() error {
	resp, err := http.Get(t.URL + "/list")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body := make([]byte, 0, 1024)
	n, curN := 0, 0
	err = nil
	for !errors.Is(err, io.EOF) {
		buf := make([]byte, 1024)
		curN, err = resp.Body.Read(buf)
		if err != nil && !errors.Is(err, io.EOF) {
			return fmt.Errorf("unable to read response body data from tracker\n%w", err)
		}
		body = append(body, buf[:curN]...)
		n = n + curN
	}
	t.Peers = new(structs.Peerlist)
	err = t.Peers.Unmarshal(body)
	if err != nil {
		return fmt.Errorf("unable to parse response body data from tracker\n%w", err)
	}
	log.Default().Printf("Found %d peers: %s\n", t.Peers.Len(), t.Peers.String())
	return nil
}

func (t *Tracker) Join() error {
	id, _ := json.Marshal(t.Identity)
	_, err := http.Post(t.URL+"/join", "application/json", bytes.NewBuffer(id))
	return err
}

func (t *Tracker) Leave() error {
	id, _ := json.Marshal(t.Identity)
	_, err := http.Post(t.URL+"/leave", "application/json", bytes.NewBuffer(id))
	return err
}
