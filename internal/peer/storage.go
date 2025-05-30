package peer

import (
	"container/ring"
	"errors"
	"fmt"
	"sync"

	"github.com/vs-ude/btml/internal/model"
)

type StorageStrategy interface {
	Decide(*KnownPeer, *model.Weights) (bool, error)
	Store(model.Weights)
	Retrieve(min int) (*model.Weights, error)
}

// Only distributes updates that are at less than twice as mature as what the
// peer was sent last time.
type QuadraticStorage struct {
	storage        map[int]*model.Weights
	quadraticSteps []int
	stepSizeCap    int
	currentMax     int
	nextStep       int
	last           *ring.Ring
	sync.Mutex
}

func NewQuadraticStorage(lastN int, stepSizeCap int) *QuadraticStorage {
	return &QuadraticStorage{
		storage:        make(map[int]*model.Weights),
		quadraticSteps: make([]int, 0),
		stepSizeCap:    stepSizeCap,
		currentMax:     0,
		nextStep:       2,
		last:           ring.New(lastN),
	}
}

func (h *QuadraticStorage) Decide(p *KnownPeer, w *model.Weights) (bool, error) {
	if w.GetAge() > 4 && p.LastUpdatedAge < (w.GetAge()/2) {
		return false, nil
	}
	return true, nil
}

// Store ignores updates that are older than the current maximum age.
func (h *QuadraticStorage) Store(w model.Weights) {
	a := w.GetAge()
	// Should be strictly increasing anyway
	if a < h.currentMax {
		return
	}
	h.Lock()
	defer h.Unlock()
	// This now points to the oldest weight we stored in the ring
	h.last = h.last.Next()

	if a == h.nextStep {
		h.quadraticSteps = append(h.quadraticSteps, a)
		h.storage[a] = &w
		h.progressStep()
	} else if a > h.nextStep {
		oldW, ok := h.last.Value.(*model.Weights)
		if ok {
			h.quadraticSteps = append(h.quadraticSteps, oldW.GetAge())
			h.storage[a] = oldW
		}
		h.progressStep()
	}

	// Always store the last N weights
	h.last.Value = &w
	h.currentMax = a
}

func (h *QuadraticStorage) progressStep() {
	h.nextStep = min(2*h.nextStep, h.nextStep+h.stepSizeCap)
}

// Retrieve retrieves the last stored weights.
func (h *QuadraticStorage) Retrieve(min int) (*model.Weights, error) {
	if min >= h.currentMax {
		return nil, errors.New("already up to date")
	}
	if h.last.Value == nil {
		return nil, errors.New("no weigths stored yet")
	}
	h.Lock()
	defer h.Unlock()

	candidate := h.getOldest()
	if candidate != nil && min >= candidate.Value.(*model.Weights).GetAge() {
		// Search in the ring of most recent updates
		for candidate.Value != nil {
			if candidate.Value.(*model.Weights).GetAge() >= min {
				return candidate.Value.(*model.Weights), nil
			}
			if candidate == h.last {
				break
			}
			candidate = candidate.Next()
		}
	}
	if len(h.quadraticSteps) > 0 && min <= h.quadraticSteps[len(h.quadraticSteps)-1] {
		// Search in the list of quadratically distanced updates
		for _, step := range h.quadraticSteps {
			if step >= min {
				return h.storage[step], nil
			}
		}
	}
	return nil, errors.New("no suitable weight found")
}

// getOldest returns the container of the oldest stored weights, skipping empty
// places in the ring.
func (h *QuadraticStorage) getOldest() *ring.Ring {
	candidate := h.last.Next()
	start := candidate
	for _, ok := candidate.Value.(*model.Weights); candidate.Value == nil || !ok; {
		candidate = candidate.Next()
		if candidate == start {
			return nil
		}
	}
	return candidate
}

func (h *QuadraticStorage) String() string {
	h.Lock()
	defer h.Unlock()
	ring := ""
	if h.last.Value == nil {
		ring = "empty"
	} else {
		cur := h.last
		end := cur
		for cur.Value != nil {
			ring = fmt.Sprintf(" %d%s", cur.Value.(*model.Weights).GetAge(), ring)
			cur = cur.Prev()
			if cur == end {
				break
			}
		}
	}

	steps := ""
	if len(h.quadraticSteps) == 0 {
		steps = "empty"
	} else {
		steps = fmt.Sprintf("%v", h.quadraticSteps)
	}
	return fmt.Sprintf(
		"HalfDistanceDistribution{currentMax: %d, nextStep: %d, ring:%s, steps: %s}",
		h.currentMax, h.nextStep, ring, steps)
}
