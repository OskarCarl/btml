package model

import (
	"errors"
)

type Weights struct {
	data []byte
	age  int
}

func (w *Weights) Get() []byte {
	return w.data
}

func (w *Weights) setAge(age int) {
	w.age = age
}

func (w *Weights) GetAge() int {
	return w.age
}

func NewWeights(d []byte) *Weights {
	return &Weights{data: d}
}

type Metrics struct {
	acc, loss float32
	age       int
}

func NewMetrics(acc, loss float32) (*Metrics, error) {
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

func (m *Metrics) GetAge() int {
	return m.age
}
