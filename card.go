package fsrs

import "time"

// Card represents a flashcard in the FSRS system.
type Card struct {
	ID         int64      `json:"id"`
	State      State      `json:"state"`
	Step       int        `json:"step"`
	Stability  float64    `json:"stability"`
	Difficulty float64    `json:"difficulty"`
	Due        time.Time  `json:"due"`
	LastReview *time.Time `json:"last_review"` // Nullable time.Time
}

func (c *Card) Duplicate() *Card {
	return &Card{
		ID:         c.ID,
		State:      c.State,
		Step:       c.Step,
		Stability:  c.Stability,
		Difficulty: c.Difficulty,
		Due:        c.Due,
		LastReview: c.LastReview,
	}
}

func NewEmptyCard(id int64) *Card {
	now := time.Now().UTC()
	return &Card{
		ID:    id,
		State: Learning,
		Due:   now,
	}
}
