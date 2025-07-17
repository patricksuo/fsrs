package fsrs

// State represents the learning state of a Card.
type State int

const (
	Learning   State = iota + 1 // 1
	Review                      // 2
	Relearning                  // 3
)
