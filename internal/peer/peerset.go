package peer

import (
	"container/list"
	"errors"
	"fmt"
	"maps"
	"math"
	"slices"
	"sort"
	"sync"
	"time"

	"github.com/vs-ude/btml/internal/structs"
	"github.com/vs-ude/btml/internal/telemetry"
)

type ErrPeerInactive error

type PeerSet struct {
	unchoked       map[string]*KnownPeer // subset of known
	known          map[string]*KnownPeer
	archive        map[string]*KnownPeer
	maxSize        int
	softMaxSize    int
	archiveAfter   time.Duration
	orderedByScore *list.List
	telemetry      *telemetry.Client
	sync.Mutex
}

func NewPeerSet(size int, archiveAfter time.Duration, telemetry *telemetry.Client) *PeerSet {
	return &PeerSet{
		unchoked:       make(map[string]*KnownPeer, size),
		known:          make(map[string]*KnownPeer),
		archive:        make(map[string]*KnownPeer),
		maxSize:        size,
		softMaxSize:    int(math.Round(float64(size) / 3 * 2)),
		archiveAfter:   archiveAfter,
		orderedByScore: list.New(),
		telemetry:      telemetry,
	}
}

func (ps *PeerSet) Add(p *structs.Peer) error {
	ps.Lock()
	defer ps.Unlock()
	switch status, err := ps.CheckPeer(p); {
	case err != nil:
		return err
	case status == CHOKED:
		ps.known[p.Name].Update(p)
		if ps.Space() > 0 {
			ps.Unchoke(p.Name)
			return nil
		}
		return fmt.Errorf("peer is known and choked")
	case status == UNCHOKED:
		ps.known[p.Name].Update(p)
	case status == UNKNOWN:
		ps.known[p.Name] = NewKnownPeer(p, ps.telemetry)
		ps.known[p.Name].updateScorePropagationFunc = ps.UpdateScore
		ps.orderedByScore.PushBack(ps.known[p.Name])
		if ps.Space() > 0 {
			ps.unchoke(p.Name)
		}
	}
	return nil
}

func (ps *PeerSet) GetUnchoked() map[string]*KnownPeer {
	return ps.known
}

// Len returns the number of known peers in the set.
func (ps *PeerSet) Len() int {
	return len(ps.known)
}

// UnchokedLen returns the number of unchoked peers in the set.
func (ps *PeerSet) UnchokedLen() int {
	return len(ps.unchoked)
}

// Space returns the number of open slots that are left in the set, based on
// the hard limit.
func (ps *PeerSet) Space() int {
	return ps.maxSize - len(ps.unchoked)
}

// GetWorstUnchoked searches for the n worst unchoked peers by score. The
// returned list is sorted by score in ascending order.
// If n > len(ps.Active), it returns all active peers.
func (ps *PeerSet) GetWorstUnchoked(n int) []*KnownPeer {
	keys := make([]*KnownPeer, 0, min(n, len(ps.unchoked)))
	for p := range maps.Values(ps.unchoked) {
		keys = append(keys, p)
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i].score < keys[j].score
	})
	return keys[:min(n, len(keys))]
}

// GetBestChoked searches for the n best choked peers by score. The
// returned list is sorted by score in descending order.
// Returns at most all choked peers, if their amount is <= n.
func (ps *PeerSet) GetBestChoked(n int) []*KnownPeer {
	keys := make([]*KnownPeer, 0, n)
	for e := ps.orderedByScore.Front(); e != nil && len(keys) < n; e = e.Next() {
		if e.Value.(*KnownPeer).State == CHOKED {
			keys = append(keys, e.Value.(*KnownPeer))
		}
	}
	return keys
}

// MultiChoke chokes the n worst-scoring peers.
// Overwrites peers in the Choked set if necessary.
func (ps *PeerSet) MultiChoke(n int) {
	ps.Lock()
	defer ps.Unlock()
	for _, p := range ps.GetWorstUnchoked(n) {
		ps.choke(p.Name)
	}
}

// Choke a single peer by name. If there was no contact with the peer in the
// last `PeerSet.archiveAfter` duration, it is archived.
func (ps *PeerSet) Choke(p string) {
	ps.Lock()
	defer ps.Unlock()
	ps.choke(p)
}

func (ps *PeerSet) choke(p string) {
	ps.known[p].choke()
	delete(ps.unchoked, p)
	if ps.known[p].LastSeen.Before(time.Now().Add(-ps.archiveAfter)) {
		ps.archive[p] = ps.known[p]
		ps.archive[p].State = ARCHIVED
		delete(ps.known, p)
	}
}

// Unchoke a single peer by name.
func (ps *PeerSet) Unchoke(p string) error {
	ps.Lock()
	defer ps.Unlock()
	return ps.unchoke(p)
}

func (ps *PeerSet) unchoke(p string) error {
	if _, ok := ps.unchoked[p]; ok {
		return nil
	}
	if len(ps.unchoked) == ps.maxSize {
		return errors.New("max amount of unchoked peers reached")
	}
	ps.unchoked[p] = ps.known[p]
	ps.unchoked[p].unchoke()

	return nil
}

// CheckPeer verifies whether the given peer is new or if it is a legitimate replacement for a known one.
func (ps *PeerSet) CheckPeer(new *structs.Peer) (peerStatus, error) {
	if _, ok := ps.unchoked[new.Name]; ok {
		// TODO: properly verify the fingerprint
		if ps.known[new.Name].Fingerprint == new.Fingerprint {
			return UNCHOKED, nil
		} else {
			return ERR, fmt.Errorf("unchoked peer exists and the new one has a non-matching fingerprint")
		}
	}
	if p, ok := ps.known[new.Name]; ok {
		// TODO: properly verify the fingerprint and check the score
		if p.Fingerprint == new.Fingerprint {
			return CHOKED, nil
		} else {
			return ERR, fmt.Errorf("choked peer exists and the new one has a non-matching fingerprint")
		}
	}
	if p, ok := ps.archive[new.Name]; ok {
		if p.Fingerprint == new.Fingerprint {
			return ARCHIVED, nil
		} else {
			return ERR, fmt.Errorf("archived peer exists and the new one has a non-matching fingerprint")
		}
	}
	return UNKNOWN, nil
}

// UpdateScore updates the internal data structures to reflect the new score of a peer.
func (ps *PeerSet) UpdateScore(kp *KnownPeer) error {
	ps.Lock()
	defer ps.Unlock()
	return ps.updateScore(kp)
}

func (ps *PeerSet) updateScore(kp *KnownPeer) error {
	var kp_element *list.Element
	var mark *list.Element
	found_kp, found_mark := false, false
	for e := ps.orderedByScore.Front(); e != nil && !(found_kp && found_mark); e = e.Next() {
		if e.Value.(*KnownPeer).Name == kp.Name {
			kp_element = e
			found_kp = true
			continue
		}
		if e.Value.(*KnownPeer).score < kp.score {
			found_mark = true
		} else {
			mark = e
		}
	}
	if kp_element == nil {
		return fmt.Errorf("peer not found")
	}
	if mark != nil {
		ps.orderedByScore.MoveAfter(kp_element, mark)
	} else {
		ps.orderedByScore.MoveToFront(kp_element)
	}
	return nil
}

func (ps *PeerSet) UnchokedToString() []string {
	keys := make([]string, 0, len(ps.unchoked))
	for k := range ps.unchoked {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	return keys
}
