package structs_test

import (
	"net"
	"testing"
	"time"

	s "github.com/vs-ude/btml/internal/structs"
)

func TestMashallingReverse(t *testing.T) {
	// prepare
	input := buildInput()

	// run
	b, err := input.Marshal()
	if err != nil {
		t.Error("unable to Marshal peerlist")
	}
	output := new(s.Peerlist)
	err = output.Unmarshal(b)
	if err != nil {
		t.Error("unable to Unmarshal peerlist")
	}

	// verify
	for key, peer := range output.List {
		expect := input.List[key]
		if !equal(t, peer, expect) {
			t.Error("peerlists are not equal")
		}
	}
	for key, peer := range input.List {
		expect := output.List[key]
		if !equal(t, peer, expect) {
			t.Error("peerlists are not equal")
		}
	}
}

func TestUnmarshalShouldError(t *testing.T) {
	// prepare
	pl := buildInput()
	b, _ := pl.Marshal()
	b[0] = 0x00

	// run
	err := pl.Unmarshal(b)

	// verify
	if err == nil {
		t.Error("unmarshal did not throw an error")
	}
}

func buildInput() *s.Peerlist {
	addr1, _ := net.ResolveUDPAddr("udp", ":43439")
	addr2, _ := net.ResolveUDPAddr("udp", "localhost:62123")
	time1, _ := time.Parse("yyyy-mm-dd hh:mm", "1999-12-31 23:59")
	time2, _ := time.Parse("yyyy-mm-dd hh:mm", "2000-01-01 00:00")
	a := &s.Peer{
		Name:        "a",
		Addr:        addr1,
		Fingerprint: "akljsdh",
		LastSeen:    time1,
	}
	b := &s.Peer{
		Name:        "b",
		Addr:        addr2,
		Fingerprint: "lkjajf",
		LastSeen:    time2,
	}
	return &s.Peerlist{
		List: map[string]*s.Peer{
			"a": a,
			"b": b,
		},
	}
}

func equal(t *testing.T, a, b *s.Peer) bool {
	ret := true
	if a.Name != b.Name {
		t.Logf("Names do not match: %s != %s", a.Name, b.Name)
		ret = false
	}
	if a.Fingerprint != b.Fingerprint {
		t.Logf("Fingerprints do not match: %s != %s", a.Fingerprint, b.Fingerprint)
		ret = false
	}
	if a.Addr.String() != b.Addr.String() {
		t.Logf("Addrs do not match: %v != %v", a.Addr.String(), b.Addr.String())
		ret = false
	}
	if a.Addr.Network() != b.Addr.Network() {
		t.Logf("Addr.Networks do not match: %s != %s", a.Addr.Network(), b.Addr.Network())
		ret = false
	}
	return ret
}
