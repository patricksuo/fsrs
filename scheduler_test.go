package fsrs

import (
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestReviewCard(t *testing.T) {
	scheduler := DefaultScheduler()
	scheduler.EnableFuzzing = false

	ratingList := []Rating{Good,
		Good,
		Good,
		Good,
		Good,
		Good,
		Again,
		Again,
		Good,
		Good,
		Good,
		Good,
		Good,
	}

	var ivlHistory []int
	card := NewEmptyCard(1)
	card.LastReview = &card.Due

	reivewDatetime := time.Date(2022, 11, 29, 12, 30, 0, 0, time.UTC)

	for _, r := range ratingList {
		card = scheduler.ReviewCard(card, r, reivewDatetime)

		ivl := int((math.Round(card.Due.Sub(*card.LastReview).Hours() / 24)))
		ivlHistory = append(ivlHistory, ivl)

		reivewDatetime = card.Due
	}

	assert.Equal(t, []int{0,
		4,
		14,
		45,
		135,
		372,
		0,
		0,
		2,
		5,
		10,
		20,
		40}, ivlHistory)
}

func TestRepeatedCorrectReviews(t *testing.T) {
	scheduler := DefaultScheduler()
	scheduler.EnableFuzzing = false

	card := NewEmptyCard(1)
	card.LastReview = &card.Due

	assert.GreaterOrEqual(t, time.Now().UTC(), card.Due)
}

func TestRetrievability(t *testing.T) {
	scheduler := DefaultScheduler()
	card := NewEmptyCard(1)

	// Retrievability of New card
	assert.Equal(t, Learning, card.State)
	retrievability := scheduler.GetCardRetrievability(card, time.Now().UTC())
	assert.Equal(t, float64(0), retrievability)

	// Retrievability of Learning card
	card = scheduler.ReviewCard(card, Good, time.Now().UTC())
	assert.Equal(t, Learning, card.State)
	retrievability = scheduler.GetCardRetrievability(card, time.Now().UTC())
	assert.GreaterOrEqual(t, retrievability, float64(0))
	assert.LessOrEqual(t, retrievability, float64(1))

	// Retrievability of Review card
	card = scheduler.ReviewCard(card, Good, time.Now().UTC())
	assert.Equal(t, Review, card.State)
	retrievability = scheduler.GetCardRetrievability(card, time.Now().UTC())
	assert.GreaterOrEqual(t, retrievability, float64(0))
	assert.LessOrEqual(t, retrievability, float64(1))

	// Retrievability of Relearning card
	card = scheduler.ReviewCard(card, Again, time.Now().UTC())
	assert.Equal(t, Relearning, card.State)
	retrievability = scheduler.GetCardRetrievability(card, time.Now().UTC())
	assert.GreaterOrEqual(t, retrievability, float64(0))
	assert.LessOrEqual(t, retrievability, float64(1))
}

func TestGoodLearningSteps(t *testing.T) {
	scheduler := DefaultScheduler()
	createdAt := time.Now().UTC()
	card := NewEmptyCard(1)

	// Initial state
	assert.Equal(t, Learning, card.State)
	assert.Equal(t, 0, card.Step)

	// First review with Good rating
	rating := Good
	card = scheduler.ReviewCard(card, rating, card.Due)
	assert.Equal(t, Learning, card.State)
	assert.Equal(t, 1, card.Step)
	// card is due in approx. 10 minutes (600 seconds)
	assert.Equal(t, 10, int(math.Round(float64(card.Due.Sub(createdAt).Minutes()))))

	// Second review with Good rating
	rating = Good
	card = scheduler.ReviewCard(card, rating, card.Due)
	assert.Equal(t, Review, card.State)
	assert.Equal(t, -1, card.Step)
	// card is due in over a day
	assert.GreaterOrEqual(t, int(card.Due.Sub(createdAt).Hours()), 24)
}

func TestAgainLearningSteps(t *testing.T) {
	scheduler := DefaultScheduler()
	createdAt := time.Now().UTC()
	card := NewEmptyCard(1)

	// Initial state
	assert.Equal(t, Learning, card.State)
	assert.Equal(t, 0, card.Step)

	// review with Again rating
	rating := Again
	card = scheduler.ReviewCard(card, rating, card.Due)
	assert.Equal(t, Learning, card.State)
	assert.Equal(t, 0, card.Step)
	// card is due in approx. 1 minute (60 seconds)
	assert.Equal(t, 1, int(math.Round(float64(card.Due.Sub(createdAt).Minutes()))))
}

func TestHardLearningSteps(t *testing.T) {
	scheduler := DefaultScheduler()
	createdAt := time.Now().UTC()
	card := NewEmptyCard(1)

	// Initial state
	assert.Equal(t, Learning, card.State)
	assert.Equal(t, 0, card.Step)

	// review with Again rating
	rating := Hard
	card = scheduler.ReviewCard(card, rating, card.Due)
	assert.Equal(t, Learning, card.State)
	assert.Equal(t, 0, card.Step)
	// card is due in approx. 5.5 minutes (330 seconds)
	assert.Equal(t, 33, int(math.Round(float64(card.Due.Sub(createdAt).Seconds()/10))))
}

func TestEasyLearningSteps(t *testing.T) {
	scheduler := DefaultScheduler()
	createdAt := time.Now().UTC()
	card := NewEmptyCard(1)

	// Initial state
	assert.Equal(t, Learning, card.State)
	assert.Equal(t, 0, card.Step)

	// review with Again rating
	rating := Easy
	card = scheduler.ReviewCard(card, rating, card.Due)
	assert.Equal(t, Review, card.State)
	assert.Equal(t, -1, card.Step)
	// card is due in at least 1 full day
	assert.GreaterOrEqual(t, int(card.Due.Sub(createdAt).Hours()), 24)
}

func TestReviewState(t *testing.T) {
	scheduler := DefaultScheduler()
	scheduler.EnableFuzzing = false
	card := NewEmptyCard(1)

	// First review with Good rating
	rating := Good
	card = scheduler.ReviewCard(card, rating, card.Due)

	// Second review with Good rating
	rating = Good
	card = scheduler.ReviewCard(card, rating, card.Due)

	assert.Equal(t, Review, card.State)
	assert.Equal(t, -1, card.Step)

	// Third review with Good rating
	prevDue := card.Due
	rating = Good
	card = scheduler.ReviewCard(card, rating, card.Due)

	assert.Equal(t, Review, card.State)
	assert.GreaterOrEqual(t, int(math.Round(card.Due.Sub(prevDue).Hours())), 24)

	// Fourth review with Again rating
	prevDue = card.Due
	rating = Again
	card = scheduler.ReviewCard(card, rating, card.Due)

	t.Logf("\nWTF %v %v %v\n", prevDue, card.Due, math.Round(card.Due.Sub(prevDue).Minutes()))
	assert.Equal(t, Relearning, card.State)

	assert.Equal(t, 10, int(math.Round(card.Due.Sub(prevDue).Minutes())))
}

func TestRelearning(t *testing.T) {
	scheduler := DefaultScheduler()
	scheduler.EnableFuzzing = false
	card := NewEmptyCard(1)

	// First review with Good rating
	rating := Good
	card = scheduler.ReviewCard(card, rating, card.Due)

	// Second review with Good rating
	rating = Good
	card = scheduler.ReviewCard(card, rating, card.Due)

	// Third review with Good rating
	prevDue := card.Due
	rating = Good
	card = scheduler.ReviewCard(card, rating, card.Due)

	// Fourth review with Again rating
	prevDue = card.Due
	rating = Again
	card = scheduler.ReviewCard(card, rating, card.Due)

	assert.Equal(t, Relearning, card.State)
	assert.Equal(t, 0, card.Step)
	assert.Equal(t, 10, int(math.Round(card.Due.Sub(prevDue).Minutes())))

	// Fifth review with Again rating
	prevDue = card.Due
	rating = Again
	card = scheduler.ReviewCard(card, rating, card.Due)

	assert.Equal(t, Relearning, card.State)
	assert.Equal(t, 0, card.Step)
	assert.Equal(t, 10, int(math.Round(card.Due.Sub(prevDue).Minutes())))

	// Sixth review with Good rating
	prevDue = card.Due
	rating = Good
	card = scheduler.ReviewCard(card, rating, card.Due)

	assert.Equal(t, Review, card.State)
	assert.Equal(t, -1, card.Step)
	assert.GreaterOrEqual(t, int(math.Round(card.Due.Sub(prevDue).Hours())), 24)
}
