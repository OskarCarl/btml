package peer

import (
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/vs-ude/btml/internal/structs"
)

func TestPeerSetUpdateScore(t *testing.T) {
	// prepare
	ps := buildPeerSet(5)
	ps.known["peer0"].score = 20
	ps.known["peer1"].score = 10
	ps.known["peer2"].score = 30
	ps.known["peer3"].score = 15
	ps.known["peer4"].score = 25
	err := make([]error, 5)

	// run
	err[0] = ps.UpdateScore(ps.known["peer0"])
	err[1] = ps.UpdateScore(ps.known["peer1"])
	err[2] = ps.UpdateScore(ps.known["peer2"])
	err[3] = ps.UpdateScore(ps.known["peer3"])
	err[4] = ps.UpdateScore(ps.known["peer4"])

	// verify
	for e := range err {
		assert.NoError(t, err[e], "error updating score")
	}
	assert.Equal(t, "peer1", ps.orderedByScore.Back().Value.(*KnownPeer).Name, "peer1 should have the lowest score")
	assert.Equal(t, "peer2", ps.orderedByScore.Front().Value.(*KnownPeer).Name, "peer2 should have the highest score")
}

func TestPeerSetUpdateScoreIndirect(t *testing.T) {
	// prepare
	ps := buildPeerSet(5)

	// run
	ps.known["peer0"].UpdateScore(20)
	ps.known["peer1"].UpdateScore(10)
	ps.known["peer2"].UpdateScore(30)
	ps.known["peer3"].UpdateScore(15)
	ps.known["peer4"].UpdateScore(25)

	// verify
	assert.Equal(t, "peer1", ps.orderedByScore.Back().Value.(*KnownPeer).Name, "peer1 should have the lowest score")
	assert.Equal(t, "peer2", ps.orderedByScore.Front().Value.(*KnownPeer).Name, "peer2 should have the highest score")
}

func TestShouldErrorForUnkownPeerScoreUpdate(t *testing.T) {
	// prepare
	ps := buildPeerSet(0)
	invalidPeer := NewKnownPeer(&structs.Peer{Name: "peer10"}, nil)

	// run
	should_error := ps.UpdateScore(invalidPeer)

	// verify
	assert.Error(t, should_error, "no error when updating the score of an unknown peer")
}

func TestGetWorstUnchoked(t *testing.T) {
	// prepare
	ps := buildPeerSet(8)
	for i := range len(ps.known) {
		ps.known["peer"+strconv.Itoa(i)].UpdateScore(10 - i)
	}
	const LEN = 3
	ps.Choke("peer3")
	ps.Choke("peer5")
	ps.Choke("peer7")

	// run
	lowest := ps.GetWorstUnchoked(LEN)

	// verify
	if assert.Equal(t, LEN, len(lowest)) {
		assert.Equal(t, lowest[0].Name, "peer6", "ordering of peers is incorrect")
		assert.Equal(t, lowest[1].Name, "peer4", "ordering of peers is incorrect")
		assert.Equal(t, lowest[2].Name, "peer2", "ordering of peers is incorrect")
	}
}

func TestGetBestChoked(t *testing.T) {
	// prepare
	ps := buildPeerSet(8)
	for i := range len(ps.known) {
		name := "peer" + strconv.Itoa(i)
		ps.known[name].UpdateScore(10 - i)
	}
	ps.Choke("peer2")
	ps.Choke("peer4")
	ps.Choke("peer5")
	ps.Choke("peer7")
	const LEN = 3

	// run
	lowest := ps.GetBestChoked(LEN)

	// verify
	if assert.Equal(t, LEN, len(lowest)) {
		assert.Equal(t, lowest[0].Name, "peer2", "ordering of peers is incorrect")
		assert.Equal(t, lowest[1].Name, "peer4", "ordering of peers is incorrect")
		assert.Equal(t, lowest[2].Name, "peer5", "ordering of peers is incorrect")
	}
}

func buildPeerSet(length int) *PeerSet {
	ps := NewPeerSet(length, time.Hour, nil)

	for i := range length {
		ps.Add(&structs.Peer{Name: "peer" + strconv.Itoa(i)})
	}
	return ps
}
