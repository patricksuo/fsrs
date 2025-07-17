package fsrs

import "time"

// ReviewLog represents the log entry of a Card object that has been reviewed.
type ReviewLog struct {
	CardID         int64     `json:"card_id"`
	Rating         Rating    `json:"rating"`
	ReviewDateTime time.Time `json:"review_datetime"`
	ReviewDuration *int64    `json:"review_duration"` // second
}
