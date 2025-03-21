package peer

import (
	"container/ring"
	"errors"
	"fmt"
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
type QuadraticDistribution struct {
	storage        map[int]*model.Weights
	quadraticSteps []int
	stepSizeCap    int
	currentMax     int
	nextStep       int
	last           *ring.Ring
	sync.Mutex
}

func NewQuadraticDistribution(lastN int, stepSizeCap int) *QuadraticDistribution {
	return &QuadraticDistribution{
		storage:        make(map[int]*model.Weights),
		quadraticSteps: make([]int, 0),
		stepSizeCap:    stepSizeCap,
		currentMax:     0,
		nextStep:       2,
		last:           ring.New(lastN),
	}
}

func (h *QuadraticDistribution) Decide(p *KnownPeer, w *model.Weights) (bool, error) {
	if w.GetAge() > 4 && p.LastUpdatedAge < (w.GetAge()/2) {
		return false, nil
	}
	return true, nil
}

// Store ignores updates that are older than the current maximum age.
func (h *QuadraticDistribution) Store(w model.Weights) {
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
		oldW := h.last.Value.(*model.Weights)
		h.quadraticSteps = append(h.quadraticSteps, oldW.GetAge())
		h.storage[a] = oldW
		h.progressStep()
	}

	// Always store the last N weights
	h.last.Value = &w
	h.currentMax = a
}

func (h *QuadraticDistribution) progressStep() {
	h.nextStep = min(2*h.nextStep, h.nextStep+h.stepSizeCap)
}

// Retrieve retrieves the last stored weights.
func (h *QuadraticDistribution) Retrieve(min int) (*model.Weights, error) {
	if min >= h.currentMax {
		return nil, errors.New("already up to date")
	}
	if h.last.Value == nil {
		return nil, errors.New("no weigths stored yet")
	}

	candidate := h.last.Next() // This gives us the oldest update
	if min >= candidate.Value.(*model.Weights).GetAge() {
		// Search in the ring of most recent updates
		start := candidate
		for candidate.Value != nil {
			if candidate.Value.(*model.Weights).GetAge() >= min {
				return candidate.Value.(*model.Weights), nil
			}
			candidate = candidate.Next()
			if candidate == start {
				break
			}
		}
	}
	if min <= h.quadraticSteps[len(h.quadraticSteps)-1] {
		h.Lock()
		defer h.Unlock()
		// Search in the list of quadratically distanced updates
		for _, step := range h.quadraticSteps {
			if step >= min {
				return h.storage[step], nil
			}
		}
	}
	return nil, errors.New("no suitable weight found")
}

func (h *QuadraticDistribution) String() string {
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
