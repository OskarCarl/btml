package peer

import (
	"fmt"
	"slices"

	"maps"

	"github.com/vs-ude/btml/internal/structs"
	"github.com/vs-ude/btml/internal/trust"
)

type ErrPeerInactive error

type peerStatus int

const (
	ERR peerStatus = iota
	UNCHOKED
	CHOKED
	UNKNOWN
)

type KnownPeer struct {
	S trust.Score
	P *structs.Peer
}

type PeerSet struct {
	Active,
	Choked map[string]*KnownPeer
}

func NewPeerSet() *PeerSet {
	return &PeerSet{
		Active: make(map[string]*KnownPeer),
		Choked: make(map[string]*KnownPeer),
	}
}

func (ps *PeerSet) Add(p *structs.Peer) error {
	switch status, err := ps.CheckPeer(p); {
	case err != nil:
		return err
	case status == CHOKED:
		return fmt.Errorf("peer is known and choked")
	case status == UNCHOKED || status == UNKNOWN:
		ps.Active[p.String()] = &KnownPeer{S: 0, P: p.Copy()}
	}
	return nil
}

// MultiChoke chokes the n worst-scoring peers.
// Overwrites peers in the Choked set if necessary.
func (ps *PeerSet) MultiChoke(n int) {
	if n > len(ps.Active) {
		maps.Copy(ps.Choked, ps.Active)
		clear(ps.Active)
	}

	lowest := ps.GetWorstUnchoked(n)
	for _, p := range lowest {
		ps.Choke(p)
	}
}

// Choke adds the given peer to the choke set and removes it from the active
// set.
func (ps *PeerSet) Choke(p string) {
	ps.Choked[p] = ps.Active[p]
	delete(ps.Active, p)
}

func (ps *PeerSet) GetWorstUnchoked(n int) []string {
	keys := make([]string, 0, len(ps.Active))
	for k := range ps.Active {
		keys = append(keys, k)
	}
	slices.SortStableFunc(keys, func(a, b string) int {
		return int(ps.Active[a].S - ps.Active[b].S) // sorts lowest to highest
	})
	return keys[:n]
}

func (ps *PeerSet) Unchoke(n int) {
	// TODO: implement
}

func (ps *PeerSet) GetBestChoked(n int) []string {
	// TODO: implement
	return nil
}

// CheckPeer verifies whether the given peer is new or if it is a legitimate replacement for a known one.
func (ps *PeerSet) CheckPeer(new *structs.Peer) (peerStatus, error) {
	if p, ok := ps.Active[new.String()]; ok {
		// TODO: properly verify the fingerprint
		if p.P.Fingerprint == new.Fingerprint {
			return UNCHOKED, nil
		} else {
			return ERR, fmt.Errorf("unchoked peer exists and the new one has a non-matching fingerprint")
		}
	}
	if p, ok := ps.Choked[new.String()]; ok {
		// TODO: properly verify the fingerprint and check the score
		if p.P.Fingerprint == new.Fingerprint {
			return CHOKED, nil
		} else {
			return ERR, fmt.Errorf("choked peer exists and the new one has a non-matching fingerprint")
		}
	}
	return UNKNOWN, nil
}
