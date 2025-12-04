package structs

type Weights struct {
	data []byte
	age  int
}

func (w *Weights) Get() []byte {
	return w.data
}

func (w *Weights) SetAge(age int) {
	w.age = age
}

func (w *Weights) GetAge() int {
	return w.age
}

func NewWeights(data []byte, age int) *Weights {
	return &Weights{data: data, age: age}
}
