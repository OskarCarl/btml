package peer

import (
	"container/ring"
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"github.com/vs-ude/btml/internal/model"
)

type StorageStrategy interface {
	Decide(*KnownPeer, *model.Weights) (bool, error)
	Store(model.Weights)
	Retrieve(min int) (*model.Weights, error)
}

// This strategy attempts to only serve updates that are at most double the age
// of the given min. There is some leeway in this due to the probability of
// gaps in the stored updates.
type DoubleAgeStorage struct {
	storage     map[int]*model.Weights
	steps       []int
	stepSizeCap int
	currentMax  int
	nextStep    int
	last        *ring.Ring
	sync.Mutex
}

func NewDoubleAgeStorage(lastN int, stepSizeCap int) *DoubleAgeStorage {
	return &DoubleAgeStorage{
		storage:     make(map[int]*model.Weights),
		steps:       make([]int, 0),
		stepSizeCap: stepSizeCap,
		currentMax:  0,
		nextStep:    2,
		last:        ring.New(lastN),
	}
}

func (h *DoubleAgeStorage) Decide(p *KnownPeer, w *model.Weights) (bool, error) {
	if w.GetAge() > 4 && p.LastUpdatedAge < (w.GetAge()/2) {
		return false, nil
	}
	return true, nil
}

// Store ignores updates that are older than the current maximum age.
func (h *DoubleAgeStorage) Store(w model.Weights) {
	a := w.GetAge()
	// Should be strictly increasing anyway
	if a < h.currentMax {
		slog.Info("Got updates to store in non-incremental order! Ignoring", "new", a, "currentMax", h.currentMax)
		return
	}
	h.Lock()
	defer h.Unlock()

	if a >= h.nextStep {
		oldW, ok := h.last.Value.(*model.Weights)
		if ok && oldIsCloser(oldW.GetAge(), h.nextStep, a) {
			h.steps = append(h.steps, oldW.GetAge())
			h.storage[oldW.GetAge()] = oldW
		} else {
			h.steps = append(h.steps, a)
			h.storage[a] = &w
		}
		h.progressStep(a)
	}

	// This now points to the oldest weight we stored in the ring
	h.last = h.last.Next()
	h.last.Value = &w
	h.currentMax = a
}

// oldIsCloser tests whether the older age is closer to the step than the
// newer. It assumes old < step && step <= new
func oldIsCloser(old, step, new int) bool {
	return step-old < new-step
}

func (h *DoubleAgeStorage) progressStep(c int) {
	for h.nextStep <= c {
		h.nextStep = min(2*h.nextStep, h.nextStep+h.stepSizeCap)
	}
}

// Retrieve retrieves the last stored weights.
func (h *DoubleAgeStorage) Retrieve(min int) (*model.Weights, error) {
	if min >= h.currentMax {
		return nil, errors.New("already up to date")
	}
	if h.last.Value == nil {
		return nil, errors.New("no weights stored yet")
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
	if len(h.steps) > 0 && min <= h.steps[len(h.steps)-1] {
		// Search in the list of older updates
		prev := -1
		for _, cur := range h.steps {
			if cur >= min {
				if prev != -1 && oldIsCloser(prev, min, cur) {
					return h.storage[prev], nil
				}
				return h.storage[cur], nil
			}
		}
	}
	return nil, errors.New("no suitable weight found")
}

// getOldest returns the container of the oldest stored weights, skipping empty
// places in the ring.
func (h *DoubleAgeStorage) getOldest() *ring.Ring {
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

func (h *DoubleAgeStorage) String() string {
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
	if len(h.steps) == 0 {
		steps = "empty"
	} else {
		steps = fmt.Sprintf("%v", h.steps)
	}
	return fmt.Sprintf(
		"DoubleAgeStorage{currentMax: %d, nextStep: %d, ring:%s, steps: %s}",
		h.currentMax, h.nextStep, ring, steps)
}
