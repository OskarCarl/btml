package tracker

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"log/slog"
	"math/big"
	"net"
	"net/http"
	"slices"
	"strconv"

	"github.com/vs-ude/btml/internal/structs"
)

// initPeer gives a requesting peer all information it needs to initialize itself correctly to join the swarm
func (t *Tracker) initPeer(w http.ResponseWriter, r *http.Request) {
	if t.telemetry.enabled && !t.telemetry.ready {
		http.Error(w, "telemetry not ready", http.StatusServiceUnavailable)
		return
	}
	i, err := t.getRandomUnusedPeerID()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		slog.Warn("Failed to get random unused peer ID", "error", err)
		return
	}
	host, _, _ := net.SplitHostPort(r.RemoteAddr)
	who := structs.WhoAmI{
		Id:         i,
		Dataset:    t.conf.Peer.Dataset,
		UpdateFreq: t.conf.Peer.UpdateFreq,
		ExtIp:      host,
		Telemetry:  *t.conf.TelConf,
	}
	buf, _ := json.Marshal(who)
	w.Write(buf)
}

func (t *Tracker) getRandomUnusedPeerID() (int, error) {
	if t.peers.Len() > t.conf.Tracker.MaxPeers {
		return -1, errors.New("max peers reached")
	}
	for {
		i, _ := rand.Int(rand.Reader, big.NewInt(int64(t.conf.Tracker.MaxPeers)))
		candidate := strconv.Itoa(int(i.Int64()))
		t.blockedPeerIds.Lock()
		if !slices.Contains(t.blockedPeerIds.list, candidate) {
			t.blockedPeerIds.list = append(t.blockedPeerIds.list, candidate)
			t.blockedPeerIds.Unlock()
			return int(i.Int64()), nil
		}
		t.blockedPeerIds.Unlock()
	}
}
