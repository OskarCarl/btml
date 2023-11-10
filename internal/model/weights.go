package model

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
