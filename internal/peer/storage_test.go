package peer_test

import (
	"testing"

	"github.com/vs-ude/btml/internal/model"
	"github.com/vs-ude/btml/internal/peer"
)

func TestPrintQuadraticStorage(t *testing.T) {
	d := prepareQuadraticStorage()
	t.Log(d)
}

func TestQuadraticStorageLast(t *testing.T) {
	// prepare
	d := prepareQuadraticStorage()
	requests := []int{19, 20}
	results := make([]int, 0, len(requests))

	// run
	for _, a := range requests {
		results = append(results, retrieve(t, d, a))
	}

	// verify
	for i := range len(results) {
		if results[i] == -1 {
			continue
		}
		if results[i] != requests[i] {
			t.Errorf("for age %d expected %d, got %d", requests[i], requests[i], results[i])
		}
	}
}

func TestQuadraticStorageSteps(t *testing.T) {
	// prepare
	d := prepareQuadraticStorage()
	requests := []int{1, 2, 3, 5, 11, 15, 16}
	expect := []int{2, 2, 4, 8, 14, 20, 20}
	results := make([]int, 0, len(requests))

	// run
	for _, a := range requests {
		results = append(results, retrieve(t, d, a))
	}

	// verify
	for i := range len(results) {
		if results[i] == -1 {
			continue
		}
		if results[i] != expect[i] {
			t.Errorf("for age %d expected %d, got %d", requests[i], expect[i], results[i])
		}
	}
}

func prepareQuadraticStorage() peer.StorageStrategy {
	d := peer.NewQuadraticStorage(3, 6)

	for a := range 22 {
		d.Store(*model.NewWeights(make([]byte, 0), a))
	}

	return d
}

func retrieve(t *testing.T, d peer.StorageStrategy, min int) int {
	w, err := d.Retrieve(min)
	if err != nil {
		t.Errorf("error retrieving weights for %d: %v", min, err)
		return -1
	}
	return w.GetAge()
}
