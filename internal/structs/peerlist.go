package structs

import (
	"encoding/json"
	"fmt"
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
	pl.List[p.String()] = p
	pl.Unlock()
}

func (pl *Peerlist) Remove(p *Peer) {
	pl.Lock()
	delete(pl.List, p.String())
	pl.Unlock()
}

func (pl *Peerlist) String() string {
	return fmt.Sprintf("%v", pl.List)
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
// The list of peers is only persisted if unmarshalling was successful.
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
