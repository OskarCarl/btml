package peer

import (
	"container/ring"
	"errors"
	"sync"

	"github.com/vs-ude/btml/internal/model"
)

type DistributionStrategy interface {
	Decide(*KnownPeer, *model.Weights) (bool, error)
	Store(model.Weights)
	Retrieve(min int) (*model.Weights, error)
}

// Only distributes updates that are at less than twice as mature as what the
// peer was sent last time.
type HalfDistanceDistribution struct {
	storage          map[int]*model.Weights
	exponentialSteps []int
	stepSizeCap      int
	currentMax       int
	previousStep     int
	last             *ring.Ring
	sync.Mutex
}

func NewHalfDistanceDistribution(lastN int, stepSizeCap int) *HalfDistanceDistribution {
	return &HalfDistanceDistribution{
		storage:          make(map[int]*model.Weights),
		exponentialSteps: make([]int, 0),
		stepSizeCap:      stepSizeCap,
		currentMax:       0,
		previousStep:     0,
		last:             ring.New(lastN),
	}
}

func (h *HalfDistanceDistribution) Decide(p *KnownPeer, w *model.Weights) (bool, error) {
	if w.GetAge() > 4 && p.LastUpdatedAge < (w.GetAge()/2) {
		return false, nil
	}
	return true, nil
}

// Store ignores updates that are older than the current maximum age.
func (h *HalfDistanceDistribution) Store(w model.Weights) {
	a := w.GetAge()
	// Should be strictly increasing anyway
	if a < h.currentMax {
		return
	}
	// Store the last N weights
	h.Lock()
	defer h.Unlock()
	h.last = h.last.Next()
	h.last.Value = &w

	h.storage[a] = &w
	// If we are between two exponential steps we only keep the highest age
	if a < min(h.previousStep+h.stepSizeCap, 2*h.previousStep) {
		delete(h.storage, h.currentMax)
	} else {
		h.previousStep = a
	}
	h.currentMax = a
}

// Retrieve retrieves the last stored weights.
func (h *HalfDistanceDistribution) Retrieve(min int) (*model.Weights, error) {
	if min >= h.previousStep {
		return nil, errors.New("already close enough")
	}
	h.Lock()
	defer h.Unlock()
	for _, step := range h.exponentialSteps {
		if step >= min {
			return h.storage[step], nil
		}
	}
	// TODO traverse the ring to find the smallest fitting weight
	return nil, errors.New("no suitable weight found")
}
