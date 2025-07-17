package fsrs

// Rating represents the four possible ratings when reviewing a card.
type Rating int

const (
	Again Rating = iota + 1 // 1
	Hard                    // 2
	Good                    // 3
	Easy                    // 4
)
