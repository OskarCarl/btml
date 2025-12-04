package model

import (
	"errors"
)

type metrics struct {
	acc, loss float32
	age       int
	guesses   map[int32]float32
}

func newMetrics(acc, loss float32, guesses map[int32]float32) (*metrics, error) {
	return &metrics{
		acc:     acc,
		loss:    loss,
		guesses: guesses,
	}, nil
}

func (m *metrics) getAccuracy() (float32, error) {
	if m.acc == -1 {
		return -1, errors.New("accuracy not measured")
	}
	return m.acc, nil
}

func (m *metrics) getLoss() (float32, error) {
	if m.loss == -1 {
		return -1, errors.New("loss not measured")
	}
	return m.loss, nil
}

func (m *metrics) getAge() int {
	return m.age
}
