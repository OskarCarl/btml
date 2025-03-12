package model

import "log/slog"

type ApplyStrategy interface {
	SetModel(model *Model)
	Start(<-chan Weights) error
}

type NaiveStrategy struct {
	model *Model
}

func NewNaiveStrategy(model *Model) *NaiveStrategy {
	return &NaiveStrategy{model: model}
}

func (ns *NaiveStrategy) SetModel(model *Model) {
	ns.model = model
}

func (ns *NaiveStrategy) Start(weightsChan <-chan *Weights) error {
	go func() {
		for weights := range weightsChan {
			err := ns.model.Apply(weights)
			if err != nil {
				slog.Error("Failed applying weights", "error", err)
				continue
			}
		}
	}()
	return nil
}
