package model

import (
	"errors"
	"log"
)

type Weights interface {
	Get() []byte
}

type SimpleWeights struct {
	data []byte
}

func (w *SimpleWeights) Get() []byte {
	return w.data
}

func NewSimpleWeights(d []byte) (*SimpleWeights, error) {
	return &SimpleWeights{data: d}, nil
}

type Metrics struct {
	acc, loss float32
}

func NewMetrics(acc, loss float32) (*Metrics, error) {
	log.Default().Printf("Got metrics acc: %f, loss: %f", acc, loss)
	return &Metrics{
		acc:  acc,
		loss: loss,
	}, nil
}

func (m *Metrics) GetAccuracy() (float32, error) {
	if m.acc == -1 {
		return -1, errors.New("accuracy not measured")
	}
	return m.acc, nil
}

func (m *Metrics) GetLoss() (float32, error) {
	if m.loss == -1 {
		return -1, errors.New("loss not measured")
	}
	return m.loss, nil
}
