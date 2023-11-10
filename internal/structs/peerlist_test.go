package structs_test

import (
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
		if peer != expect {
			t.Error("peerlists are not equal")
		}
	}
	for key, peer := range input.List {
		expect := output.List[key]
		if peer != expect {
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
	return &s.Peerlist{
		List: map[string]s.Peer{
			"a": {
				Name:        "a",
				Addr:        "localhost:32453",
				Proto:       s.UDP,
				Fingerprint: "akljsdh",
			},
			"b": {
				Name:        "b",
				Addr:        "localhost:62123",
				Proto:       s.TCP,
				Fingerprint: "lkjajf",
			},
		},
	}
}
