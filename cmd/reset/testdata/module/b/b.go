package b

// generate:reset
type Inner struct {
	N    int
	Tags []string
}

// Manual has its own Reset, so the generator must call it instead of
// resetting the fields.
type Manual struct{ Calls int }

func (m *Manual) Reset() { m.Calls = -1 }

// Plain has no Reset, so the generator must replace it with its zero value.
type Plain struct{ X int }
