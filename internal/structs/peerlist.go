package structs

import (
	"encoding/json"
	"fmt"
	"sync"
)

type Peerlist struct {
	list map[string]Peer
	sync.Mutex
}

func NewPeerList() *Peerlist {
	return &Peerlist{
		list: make(map[string]Peer),
	}
}

func (pl *Peerlist) Add(p *Peer) {
	pl.Lock()
	pl.list[p.String()] = *p
	pl.Unlock()
}

func (pl *Peerlist) Remove(p *Peer) {
	pl.Lock()
	delete(pl.list, p.String())
	pl.Unlock()
}

func (pl *Peerlist) String() string {
	return fmt.Sprintf("%v", pl.list)
}

func (pl *Peerlist) Len() int {
	return len(pl.list)
}

// Unmarshal parses the byte array into a Peerlist.
// The list of peers is only persisted if unmarshalling was successful.
func (pl *Peerlist) Unmarshal(b []byte) error {
	tmp := make(map[string]Peer)
	err := json.Unmarshal(b, &tmp)
	if err != nil {
		return err
	}
	pl.Lock()
	pl.list = tmp
	pl.Unlock()
	return nil
}

func (pl *Peerlist) Marshal() ([]byte, error) {
	pl.Lock()
	b, err := json.Marshal(pl.list)
	pl.Unlock()
	return b, err
}
