package peer

import (
	"errors"
	"fmt"
	"slices"

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
	S                trust.Score
	MostRecentUpdate int
	P                *structs.Peer
	State            peerStatus
	// TODO move QUIC conn here and kill it when choking?
}

func (kp *KnownPeer) unchoke() error {
	kp.State = UNCHOKED
	return nil
}

func (kp *KnownPeer) choke() error {
	kp.State = CHOKED
	return nil
}

type PeerSet struct {
	Active  map[string]*KnownPeer // subset of Known
	Known   map[string]*KnownPeer
	MaxSize int
}

func NewPeerSet(size int) *PeerSet {
	return &PeerSet{
		Active:  make(map[string]*KnownPeer, size),
		Known:   make(map[string]*KnownPeer),
		MaxSize: size,
	}
}

func (ps *PeerSet) Add(p *structs.Peer) error {
	switch status, err := ps.CheckPeer(p); {
	case err != nil:
		return err
	case status == CHOKED:
		ps.Known[p.Name].P = p.Copy()
		return fmt.Errorf("peer is known and choked")
	case status == UNCHOKED:
		ps.Known[p.Name].P = p.Copy()
	case status == UNKNOWN:
		ps.Known[p.Name] = &KnownPeer{
			S:                0,
			MostRecentUpdate: 0,
			P:                p.Copy(),
			State:            CHOKED,
		}
	}
	return nil
}

// MultiChoke chokes the n worst-scoring peers.
// Overwrites peers in the Choked set if necessary.
// func (ps *PeerSet) MultiChoke(n int) {
// 	if n > len(ps.Active) {
// 		maps.Copy(ps.Choked, ps.Active)
// 		clear(ps.Active)
// 	}

// 	lowest := ps.GetWorstUnchoked(n)
// 	for _, p := range lowest {
// 		ps.Choke(p)
// 	}
// }

// Choke adds the given peer to the choke set and removes it from the active
// set.
func (ps *PeerSet) Choke(p string) {
	ps.Known[p].choke()
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

func (ps *PeerSet) Unchoke(p string) error {
	if _, ok := ps.Active[p]; ok {
		return nil
	}
	if len(ps.Active) == ps.MaxSize {
		return errors.New("max amount of unchoked peers reached")
	}
	ps.Active[p] = ps.Known[p]
	ps.Active[p].unchoke()

	return nil
}

func (ps *PeerSet) GetBestChoked(n int) []string {
	// TODO: implement
	return nil
}

// CheckPeer verifies whether the given peer is new or if it is a legitimate replacement for a known one.
func (ps *PeerSet) CheckPeer(new *structs.Peer) (peerStatus, error) {
	if _, ok := ps.Active[new.Name]; ok {
		// TODO: properly verify the fingerprint
		if ps.Known[new.Name].P.Fingerprint == new.Fingerprint {
			return UNCHOKED, nil
		} else {
			return ERR, fmt.Errorf("unchoked peer exists and the new one has a non-matching fingerprint")
		}
	}
	if p, ok := ps.Known[new.Name]; ok {
		// TODO: properly verify the fingerprint and check the score
		if p.P.Fingerprint == new.Fingerprint {
			return CHOKED, nil
		} else {
			return ERR, fmt.Errorf("choked peer exists and the new one has a non-matching fingerprint")
		}
	}
	return UNKNOWN, nil
}
