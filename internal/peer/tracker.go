package peer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/vs-ude/btml/internal/structs"
)

type Tracker struct {
	URL        string
	Peers      *structs.Peerlist
	Identity   *structs.Peer
	UpdateFreq time.Duration
	sync.Mutex
}

func (t *Tracker) Setup(c *Config, p *structs.Peer) {
	t.Identity = p
	t.Peers = structs.NewPeerList()

	err := t.Join()
	if err != nil {
		slog.Error("Failed joining the tracker", "error", err)
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
	slog.Info("Found peers", "count", t.Peers.Len(), "peers", t.Peers.String())
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

// periodicUpdate periodically updates the peer list from the tracker.
// This has the side effect of pinging the tracker so it knows we are alive.
func (t *Tracker) periodicUpdate(wg *sync.WaitGroup, ctx context.Context) {
	defer wg.Done()
	timer := time.NewTimer(time.Second)
	errCount := 0
	waitTime := t.UpdateFreq
	for {
		select {
		case <-timer.C:
			t.Lock()
			err := t.Update()
			t.Unlock()
			if err != nil {
				errCount++
				slog.Warn("Failed updating peers from tracker", "error", err)
				if errCount >= 3 {
					waitTime = min(waitTime*2, time.Second*120)
					slog.Warn("Too many consecutive errors updating peers", "wait_time", waitTime.String(), "error_count", errCount)
					timer.Reset(waitTime)
				}
			} else {
				errCount = 0
				waitTime = t.UpdateFreq
			}
			timer.Reset(waitTime)
		case <-ctx.Done():
			return
		}
	}
}
