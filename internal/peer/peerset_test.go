package peer_test

import (
	"strconv"
	"testing"

	"github.com/vs-ude/btml/internal/peer"
	"github.com/vs-ude/btml/internal/structs"
	"github.com/vs-ude/btml/internal/trust"
)

func TestGetWorstActive(t *testing.T) {
	// prepare
	ps := buildPeerSet()
	const LEN = 3

	// run
	lowest := ps.GetWorstUnchoked(LEN)

	// verify
	if len(lowest) != LEN {
		t.Error("returned an incorrect number of peer ids")
	}
	id0correct := lowest[0] == "peer7"
	id1correct := lowest[1] == "peer6"
	id2correct := lowest[2] == "peer5"
	if !id0correct || !id1correct || !id2correct {
		t.Error("ordering of peers is incorrect")
	}
}

func buildPeerSet() *peer.PeerSet {
	buildPeer := func(s int, n string) *peer.KnownPeer {
		return &peer.KnownPeer{
			S: trust.Score(s),
			P: &structs.Peer{
				Name: n,
			},
		}
	}

	ps := peer.NewPeerSet(8)

	for i := range 8 {
		n := strconv.Itoa(i)
		p := buildPeer(10-i, n)
		ps.Active["peer"+n] = p // Score and name are sorted inversely!
		ps.Known["peer"+n] = p  // Score and name are sorted inversely!
	}

	return ps
}
