package structs

import (
	"encoding/json"
	"fmt"
	"maps"
	"slices"
	"sync"
	"time"
)

type Peerlist struct {
	List map[string]*Peer
	sync.Mutex
}

func NewPeerList() *Peerlist {
	return &Peerlist{
		List: make(map[string]*Peer),
	}
}

func (pl *Peerlist) Add(p *Peer) {
	pl.Lock()
	pl.List[p.Name] = p
	pl.Unlock()
}

func (pl *Peerlist) Remove(p *Peer) {
	pl.Lock()
	delete(pl.List, p.Name)
	pl.Unlock()
}

func (pl *Peerlist) String() string {
	sl := slices.Collect(maps.Keys(pl.List))
	slices.Sort(sl)
	return fmt.Sprintf("%v", sl)
}

func (pl *Peerlist) Len() int {
	return len(pl.List)
}

func (pl *Peerlist) Touch(p string) {
	pl.Lock()
	pl.List[p].LastSeen = time.Now()
	pl.Unlock()
}

func (pl *Peerlist) Has(p string) bool {
	_, ok := pl.List[p]
	return ok
}

// Unmarshal parses the byte array into a Peerlist.
// The list of peers is only persisted if unmarshaling was successful.
func (pl *Peerlist) Unmarshal(b []byte) error {
	tmp := make(map[string]*Peer)
	err := json.Unmarshal(b, &tmp)
	if err != nil {
		return err
	}
	pl.Lock()
	pl.List = tmp
	pl.Unlock()
	return nil
}

func (pl *Peerlist) Marshal() ([]byte, error) {
	pl.Lock()
	b, err := json.Marshal(pl.List)
	pl.Unlock()
	return b, err
}
