package peer_test

import (
	"testing"

	"github.com/vs-ude/btml/internal/model"
	"github.com/vs-ude/btml/internal/peer"
)

func TestPrintQuadraticDistribution(t *testing.T) {
	d := prepareQuadraticDistribution()
	t.Log(d)
}

func TestQuadraticDistributionLast(t *testing.T) {
	// prepare
	d := prepareQuadraticDistribution()
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

func TestQuadraticDistributionSteps(t *testing.T) {
	// prepare
	d := prepareQuadraticDistribution()
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

func prepareQuadraticDistribution() peer.DistributionStrategy {
	d := peer.NewQuadraticDistribution(3, 6)

	for a := range 22 {
		d.Store(*model.NewWeights(make([]byte, 0), a))
	}

	return d
}

func retrieve(t *testing.T, d peer.DistributionStrategy, min int) int {
	w, err := d.Retrieve(min)
	if err != nil {
		t.Errorf("error retrieving weights for %d: %v", min, err)
		return -1
	}
	return w.GetAge()
}
