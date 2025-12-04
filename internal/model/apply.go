package model

import (
	"log/slog"

	"github.com/vs-ude/btml/internal/structs"
)

type WeightsWithCallback struct {
	callback func(int)
	*structs.Weights
}

func NewWeightsWithCallback(w *structs.Weights, callback func(int)) *WeightsWithCallback {
	return &WeightsWithCallback{
		callback: callback,
		Weights:  w,
	}
}

func (wwc *WeightsWithCallback) ToWeights() *structs.Weights {
	return wwc.Weights
}

// Type checks
var _ ApplyStrategy = &NaiveStrategy{}
var _ ApplyStrategy = &SimpleActionStrategy{}

type ApplyStrategy interface {
	SetModel(model *Model)
	Start(<-chan *WeightsWithCallback) error
}

// Applies all updates it gets and does nothing else.
type NaiveStrategy struct {
	model *Model
}

func NewNaiveStrategy(model *Model) *NaiveStrategy {
	return &NaiveStrategy{model: model}
}

func (ns *NaiveStrategy) SetModel(model *Model) {
	ns.model = model
}

func (ns *NaiveStrategy) Start(weightsChan <-chan *WeightsWithCallback) error {
	go func() {
		for weights := range weightsChan {
			_, err := ns.model.Apply(weights.ToWeights())
			if err != nil {
				slog.Error("Failed applying weights", "error", err)
				continue
			}
		}
	}()
	return nil
}

// Applies all updates and updates the score slightly based on the change in
// loss.
type SimpleActionStrategy struct {
	model           *Model
	changeThreshold float32
}

func NewSimpleActionStrategy(model *Model, changeThreshold float32) *SimpleActionStrategy {
	return &SimpleActionStrategy{
		model:           model,
		changeThreshold: changeThreshold,
	}
}

func (sas *SimpleActionStrategy) SetModel(model *Model) {
	sas.model = model
}

func (sas *SimpleActionStrategy) Start(weightsChan <-chan *WeightsWithCallback) error {
	go func() {
		for weights := range weightsChan {
			change, err := sas.model.Apply(weights.ToWeights())
			if err != nil {
				slog.Error("Failed applying weights", "error", err)
				continue
			}
			var score int
			// This assumes that change is a difference in loss, i.e. >0 = bad and <0 = good
			// It also assumes that
			switch {
			case change > sas.changeThreshold:
				score = -1
			case change < -sas.changeThreshold:
				score = 1
			default:
				return
			}
			weights.callback(score)
		}
	}()
	return nil
}
