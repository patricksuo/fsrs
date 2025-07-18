package fsrs

import (
	"math"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func mustNewScheduler(options ...SchedulerOption) *Scheduler {
	s, err := NewScheduler(options...)
	if err != nil {
		panic(err)
	}

	return s
}

func TestReviewCard(t *testing.T) {
	scheduler := mustNewScheduler(WithEnableFuzzing(false))

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
	card := NewEmptyCard(1)
	card.LastReview = &card.Due

	assert.GreaterOrEqual(t, time.Now().UTC(), card.Due)
}

func TestRetrievability(t *testing.T) {
	scheduler := mustNewScheduler()
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
	scheduler := mustNewScheduler()
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
	scheduler := mustNewScheduler()
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
	scheduler := mustNewScheduler()
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
	scheduler := mustNewScheduler()
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
	scheduler := mustNewScheduler(WithEnableFuzzing(false))
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

	assert.Equal(t, Relearning, card.State)

	assert.Equal(t, 10, int(math.Round(card.Due.Sub(prevDue).Minutes())))
}

func TestRelearning(t *testing.T) {
	scheduler := mustNewScheduler(WithEnableFuzzing(false))

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

func TestFuzzing(t *testing.T) {
	scheduler := mustNewScheduler(WithRandomSource(rand.NewSource(1)))

	card := NewEmptyCard(1)
	prevDue := card.Due
	card = scheduler.ReviewCard(card, Good, time.Now().UTC())

	prevDue = card.Due
	card = scheduler.ReviewCard(card, Good, prevDue)

	prevDue = card.Due
	card = scheduler.ReviewCard(card, Good, prevDue)

	intervalDays := int(math.Round(card.Due.Sub(prevDue).Hours() / 24))
	assert.Equal(t, 20, intervalDays)
}

func TestNoLearningSteps(t *testing.T) {
	scheduler := mustNewScheduler(WithLearningSteps(nil))
	assert.Equal(t, 0, len(scheduler.learningSteps))

	card := NewEmptyCard(1)
	card = scheduler.ReviewCard(card, Again, time.Now().UTC())

	assert.Equal(t, Review, card.State)

	intervalDays := int(math.Round(card.Due.Sub(*card.LastReview).Hours() / 24))
	assert.GreaterOrEqual(t, intervalDays, 1)
}
func TestNoRelearningSteps(t *testing.T) {
	scheduler := mustNewScheduler(WithRelearningSteps(nil))
	assert.Equal(t, 0, len(scheduler.relearningSteps))

	card := NewEmptyCard(1)
	assert.Equal(t, Learning, card.State)

	card = scheduler.ReviewCard(card, Good, time.Now().UTC())
	assert.Equal(t, Learning, card.State)

	card = scheduler.ReviewCard(card, Good, card.Due)
	assert.Equal(t, Review, card.State)

	card = scheduler.ReviewCard(card, Again, card.Due)
	assert.Equal(t, Review, card.State)

	intervalDays := int(math.Round(card.Due.Sub(*card.LastReview).Hours() / 24))
	assert.GreaterOrEqual(t, 1, intervalDays)
}

func TestOneCardMultipleSchedulers(t *testing.T) {
	// Initialize schedulers with different learning and relearning steps
	schedulerWithTwoLearningSteps := mustNewScheduler(WithLearningSteps([]time.Duration{time.Minute, 10 * time.Minute}))
	schedulerWithOneLearningStep := mustNewScheduler(WithLearningSteps([]time.Duration{time.Minute}))
	schedulerWithNoLearningSteps := mustNewScheduler(WithLearningSteps(nil))

	schedulerWithTwoRelearningSteps := mustNewScheduler(WithRelearningSteps([]time.Duration{time.Minute, 10 * time.Minute}))
	schedulerWithOneRelearningStep := mustNewScheduler(WithRelearningSteps([]time.Duration{time.Minute}))
	schedulerWithNoRelearningSteps := mustNewScheduler(WithRelearningSteps(nil))

	card := NewEmptyCard(1)

	// Learning-state tests
	assert.Equal(t, 2, len(schedulerWithTwoLearningSteps.learningSteps), "Expected 2 learning steps")
	card = schedulerWithTwoLearningSteps.ReviewCard(card, Good, time.Now().UTC())
	assert.Equal(t, Learning, card.State, "Expected card state to be Learning")
	assert.Equal(t, 1, card.Step, "Expected card step to be 1")

	assert.Equal(t, 1, len(schedulerWithOneLearningStep.learningSteps), "Expected 1 learning step")
	card = schedulerWithOneLearningStep.ReviewCard(card, Again, time.Now().UTC())
	assert.Equal(t, Learning, card.State, "Expected card state to be Learning")
	assert.Equal(t, 0, card.Step, "Expected card step to be 0")

	assert.Equal(t, 0, len(schedulerWithNoLearningSteps.learningSteps), "Expected 0 learning steps")
	card = schedulerWithNoLearningSteps.ReviewCard(card, Hard, time.Now().UTC())
	assert.Equal(t, Review, card.State, "Expected card state to be Review")
	assert.Equal(t, -1, card.Step, "Expected card step to be -1")

	// Relearning-state tests
	assert.Equal(t, 2, len(schedulerWithTwoRelearningSteps.relearningSteps), "Expected 2 relearning steps")
	card = schedulerWithTwoRelearningSteps.ReviewCard(card, Again, time.Now().UTC())
	assert.Equal(t, Relearning, card.State, "Expected card state to be Relearning")
	assert.Equal(t, 0, card.Step, "Expected card step to be 0")

	card = schedulerWithTwoRelearningSteps.ReviewCard(card, Good, time.Now().UTC())
	assert.Equal(t, Relearning, card.State, "Expected card state to be Relearning")
	assert.Equal(t, 1, card.Step, "Expected card step to be 1")

	assert.Equal(t, 1, len(schedulerWithOneRelearningStep.relearningSteps), "Expected 1 relearning step")
	card = schedulerWithOneRelearningStep.ReviewCard(card, Again, time.Now().UTC())
	assert.Equal(t, Relearning, card.State, "Expected card state to be Relearning")
	assert.Equal(t, 0, card.Step, "Expected card step to be 0")

	assert.Equal(t, 0, len(schedulerWithNoRelearningSteps.relearningSteps), "Expected 0 relearning steps")
	card = schedulerWithNoRelearningSteps.ReviewCard(card, Hard, time.Now().UTC())
	assert.Equal(t, Review, card.State, "Expected card state to be Review")
	assert.Equal(t, -1, card.Step, "Expected card step to be -1")
}

func TestMaximumInterval(t *testing.T) {
	maximumInterval := 100
	scheduler := mustNewScheduler(WithMaximumInterval(maximumInterval))

	card := NewEmptyCard(1)

	// Review with Easy rating
	card = scheduler.ReviewCard(card, Easy, card.Due)
	intervalDays := int(math.Round(card.Due.Sub(*card.LastReview).Hours() / 24))
	assert.LessOrEqual(t, intervalDays, maximumInterval, "Interval should not exceed maximum_interval after Easy rating")

	// Review with Good rating
	card = scheduler.ReviewCard(card, Good, card.Due)
	intervalDays = int(math.Round(card.Due.Sub(*card.LastReview).Hours() / 24))
	assert.LessOrEqual(t, intervalDays, maximumInterval, "Interval should not exceed maximum_interval after Good rating")

	// Review with Easy rating again
	card = scheduler.ReviewCard(card, Easy, card.Due)
	intervalDays = int(math.Round(card.Due.Sub(*card.LastReview).Hours() / 24))
	assert.LessOrEqual(t, intervalDays, maximumInterval, "Interval should not exceed maximum_interval after second Easy rating")

	// Review with Good rating again
	card = scheduler.ReviewCard(card, Good, card.Due)
	intervalDays = int(math.Round(card.Due.Sub(*card.LastReview).Hours() / 24))
	assert.LessOrEqual(t, intervalDays, maximumInterval, "Interval should not exceed maximum_interval after second Good rating")
}

func TestLearningCardRateHardOneLearningStep(t *testing.T) {
	firstLearningStep := 10 * time.Minute
	schedulerWithOneLearningStep := mustNewScheduler(WithLearningSteps([]time.Duration{firstLearningStep}))

	card := NewEmptyCard(1)

	initialDueDatetime := card.Due

	card = schedulerWithOneLearningStep.ReviewCard(card, Hard, card.Due)

	assert.Equal(t, Learning, card.State, "Expected card state to be Learning")

	newDueDatetime := card.Due

	intervalLength := newDueDatetime.Sub(initialDueDatetime)
	expectedIntervalLength := time.Duration(float64(firstLearningStep) * 1.5)
	tolerance := time.Second

	assert.LessOrEqual(t, math.Abs(float64(intervalLength-expectedIntervalLength)), float64(tolerance), "Interval length should be within tolerance of expected interval")
}

func TestLearningCardRateHardSecondLearningStep(t *testing.T) {
	firstLearningStep := time.Minute
	secondLearningStep := 10 * time.Minute
	schedulerWithTwoLearningSteps := mustNewScheduler(WithLearningSteps([]time.Duration{firstLearningStep, secondLearningStep}))

	card := NewEmptyCard(1)

	assert.Equal(t, Learning, card.State, "Expected initial card state to be Learning")
	assert.Equal(t, 0, card.Step, "Expected initial card step to be 0")

	card = schedulerWithTwoLearningSteps.ReviewCard(card, Good, card.Due)

	assert.Equal(t, Learning, card.State, "Expected card state to be Learning after Good rating")
	assert.Equal(t, 1, card.Step, "Expected card step to be 1 after Good rating")

	dueDatetimeAfterFirstReview := card.Due

	card = schedulerWithTwoLearningSteps.ReviewCard(card, Hard, dueDatetimeAfterFirstReview)

	dueDatetimeAfterSecondReview := card.Due

	assert.Equal(t, Learning, card.State, "Expected card state to be Learning after Hard rating")
	assert.Equal(t, 1, card.Step, "Expected card step to be 1 after Hard rating")

	intervalLength := dueDatetimeAfterSecondReview.Sub(dueDatetimeAfterFirstReview)
	expectedIntervalLength := secondLearningStep
	tolerance := time.Second

	assert.LessOrEqual(t, math.Abs(float64(intervalLength-expectedIntervalLength)), float64(tolerance), "Interval length should be within tolerance of second learning step")
}

func TestRelearningCardRateHardOneRelearningStep(t *testing.T) {
	firstRelearningStep := 10 * time.Minute
	schedulerWithOneRelearningStep := mustNewScheduler(WithRelearningSteps([]time.Duration{firstRelearningStep}))

	card := NewEmptyCard(1)

	card = schedulerWithOneRelearningStep.ReviewCard(card, Easy, card.Due)
	assert.Equal(t, Review, card.State, "Expected card state to be Review after Easy rating")

	card = schedulerWithOneRelearningStep.ReviewCard(card, Again, card.Due)
	assert.Equal(t, Relearning, card.State, "Expected card state to be Relearning after Again rating")
	assert.Equal(t, 0, card.Step, "Expected card step to be 0 after Again rating")

	prevDueDatetime := card.Due

	card = schedulerWithOneRelearningStep.ReviewCard(card, Hard, prevDueDatetime)
	assert.Equal(t, Relearning, card.State, "Expected card state to be Relearning after Hard rating")
	assert.Equal(t, 0, card.Step, "Expected card step to be 0 after Hard rating")

	newDueDatetime := card.Due

	intervalLength := newDueDatetime.Sub(prevDueDatetime)
	expectedIntervalLength := time.Duration(float64(firstRelearningStep) * 1.5)
	tolerance := time.Second

	assert.LessOrEqual(t, math.Abs(float64(intervalLength-expectedIntervalLength)), float64(tolerance), "Interval length should be within tolerance of 1.5 times the first relearning step")
}

func TestRelearningCardRateHardTwoRelearningSteps(t *testing.T) {
	firstRelearningStep := time.Minute
	secondRelearningStep := 10 * time.Minute
	schedulerWithTwoRelearningSteps := mustNewScheduler(WithRelearningSteps([]time.Duration{firstRelearningStep, secondRelearningStep}))

	card := NewEmptyCard(1)

	// First review: Easy rating
	card = schedulerWithTwoRelearningSteps.ReviewCard(card, Easy, card.Due)
	assert.Equal(t, Review, card.State, "Expected card state to be Review after Easy rating")

	// Second review: Again rating
	card = schedulerWithTwoRelearningSteps.ReviewCard(card, Again, card.Due)
	assert.Equal(t, Relearning, card.State, "Expected card state to be Relearning after Again rating")
	assert.Equal(t, 0, card.Step, "Expected card step to be 0 after Again rating")

	prevDueDatetime := card.Due

	// Third review: Hard rating at step 0
	card = schedulerWithTwoRelearningSteps.ReviewCard(card, Hard, prevDueDatetime)
	assert.Equal(t, Relearning, card.State, "Expected card state to be Relearning after Hard rating at step 0")
	assert.Equal(t, 0, card.Step, "Expected card step to be 0 after Hard rating at step 0")

	newDueDatetime := card.Due
	intervalLength := newDueDatetime.Sub(prevDueDatetime)
	expectedIntervalLength := time.Duration(float64(firstRelearningStep+secondRelearningStep) / 2.0)
	tolerance := time.Second
	assert.LessOrEqual(t, math.Abs(float64(intervalLength-expectedIntervalLength)), float64(tolerance), "Interval length should be within tolerance of average of relearning steps")

	// Fourth review: Good rating
	card = schedulerWithTwoRelearningSteps.ReviewCard(card, Good, card.Due)
	assert.Equal(t, Relearning, card.State, "Expected card state to be Relearning after Good rating")
	assert.Equal(t, 1, card.Step, "Expected card step to be 1 after Good rating")

	prevDueDatetime = card.Due

	// Fifth review: Hard rating at step 1
	card = schedulerWithTwoRelearningSteps.ReviewCard(card, Hard, prevDueDatetime)
	assert.Equal(t, Relearning, card.State, "Expected card state to be Relearning after Hard rating at step 1")
	assert.Equal(t, 1, card.Step, "Expected card step to be 1 after Hard rating at step 1")

	newDueDatetime = card.Due
	intervalLength = newDueDatetime.Sub(prevDueDatetime)
	expectedIntervalLength = secondRelearningStep
	assert.LessOrEqual(t, math.Abs(float64(intervalLength-expectedIntervalLength)), float64(tolerance), "Interval length should be within tolerance of second relearning step")

	// Sixth review: Easy rating
	card = schedulerWithTwoRelearningSteps.ReviewCard(card, Easy, prevDueDatetime)
	assert.Equal(t, Review, card.State, "Expected card state to be Review after Easy rating")
	assert.Equal(t, -1, card.Step, "Expected card step to be -1 after Easy rating")
}
