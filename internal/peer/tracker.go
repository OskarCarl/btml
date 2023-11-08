package peer

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

type Tracker struct {
	URL        string
	Peers      []string
	ListenAddr string
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
	data := make(map[string]bool)
	err = json.Unmarshal(body, &data)
	if err != nil {
		return fmt.Errorf("unable to parse response body data from tracker\n%w", err)
	}
	t.Peers = make([]string, 0, 100)
	for p := range data {
		if p == t.ListenAddr {
			continue
		}
		t.Peers = append(t.Peers, p)
	}
	fmt.Printf("Found %d peers: %v\n", len(t.Peers), t.Peers)
	return nil
}

func (t *Tracker) Join() error {
	_, err := http.Post(t.URL+"/join", "application/json", bytes.NewBufferString("{\"addr\": \""+t.ListenAddr+"\"}"))
	return err
}

func (t *Tracker) Leave() error {
	_, err := http.Post(t.URL+"/leave", "application/json", bytes.NewBufferString("{\"addr\": \""+t.ListenAddr+"\"}"))
	return err
}
