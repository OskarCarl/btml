package structs_test

import (
	"net"
	"testing"

	s "github.com/vs-ude/btfl/internal/structs"
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
	a := &s.Peer{
		Name:        "a",
		Addr:        addr1,
		Fingerprint: "akljsdh",
	}
	b := &s.Peer{
		Name:        "b",
		Addr:        addr2,
		Fingerprint: "lkjajf",
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
	if !a.Addr.IP.Equal(b.Addr.IP) {
		t.Logf("Addr.IPs do not match: %v != %v", a.Addr.IP, b.Addr.IP)
		ret = false
	}
	if a.Addr.Port != b.Addr.Port {
		t.Logf("Addr.Ports do not match: %d != %d", a.Addr.Port, b.Addr.Port)
		ret = false
	}
	if a.Addr.Zone != b.Addr.Zone {
		t.Logf("Addr.Zones do not match: %s != %s", a.Addr.Zone, b.Addr.Zone)
		ret = false
	}
	return ret
}
