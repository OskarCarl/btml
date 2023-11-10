package trust

type Score int

var (
	MaxScore Score = 100
	MinScore Score = 0
)

// Increments the Score up to a ceiling of MaxScore
func (s *Score) Increment(i int) {
	*s += Score(i)
	if *s > MaxScore {
		*s = MaxScore
	}
}

// Decrements the Score down to a floor of MinScore
func (s *Score) Decrement(i int) {
	*s -= Score(i)
	if *s < MinScore {
		*s = MinScore
	}
}
