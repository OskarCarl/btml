package model

import "github.com/vs-ude/btfl/internal/trust"

type Model interface {
	Eval(Weights) (trust.Score, error)
	Apply(Weights) error
	GetWeights() Weights
}
