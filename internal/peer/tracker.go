package peer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/vs-ude/btfl/internal/structs"
)

type Tracker struct {
	URL        string
	Peers      *structs.Peerlist
	Identity   *structs.Peer
	UpdateFreq time.Duration
}

func (t *Tracker) Setup(c *Config, p *structs.Peer) {
	t.Identity = p
	t.Peers = structs.NewPeerList()

	err := t.Join()
	if err != nil {
		log.Default().Printf("Error joining the tracker: %v\n", err)
		return
	}
}

func (t *Tracker) Update() error {
	req, err := http.NewRequest("GET", t.URL+"/list", http.NoBody)
	req.Header.Add("peer-id", t.Identity.Name)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := getResponseBody(resp)
	if err != nil {
		return err
	}
	t.Peers = new(structs.Peerlist)
	err = t.Peers.Unmarshal(*body)
	if err != nil {
		return fmt.Errorf("unable to parse update response body data from tracker\n%w", err)
	}
	delete(t.Peers.List, t.Identity.Name)
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

func getResponseBody(resp *http.Response) (*[]byte, error) {
	var b []byte
	if resp.ContentLength > 0 {
		b = make([]byte, 0, resp.ContentLength)
	} else {
		b = make([]byte, 0)
	}
	buf := bytes.NewBuffer(b)
	_, err := buf.ReadFrom(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read response body data from tracker\n%w", err)
	}
	body := buf.Bytes()
	return &body, nil
}
