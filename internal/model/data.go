package model

import (
	"errors"
)

type Weights interface {
	Get() []byte
	GetAge() int
	setAge(int)
}

type SimpleWeights struct {
	data []byte
	age  int
}

func (w *SimpleWeights) Get() []byte {
	return w.data
}

func (w *SimpleWeights) setAge(age int) {
	w.age = age
}

func (w *SimpleWeights) GetAge() int {
	return w.age
}

func NewSimpleWeights(d []byte) Weights {
	return &SimpleWeights{data: d}
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
