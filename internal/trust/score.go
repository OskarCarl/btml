package trust

type Score int

var (
	MaxScore Score = 100
	MinScore Score = 0
)

// Increments the Score up to a ceiling of MaxScore
func (s *Score) Increment(i int) {
	*s = min(*s+Score(i), MaxScore)
}

// Decrements the Score down to a floor of MinScore
func (s *Score) Decrement(i int) {
	*s = max(*s-Score(i), MinScore)
}
