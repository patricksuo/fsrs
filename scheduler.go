package fsrs

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
	"time"
)

type Scheduler struct {
	// Parameters are the model weights of the FSRS scheduler.
	Parameters []float64
	// DesiredRetention is the desired retention rate of cards scheduled with the scheduler.
	DesiredRetention float64
	// LearningSteps are small time intervals that schedule cards in the Learning state.
	LearningSteps []time.Duration
	// RelearningSteps are small time intervals that schedule cards in the Relearning state.
	RelearningSteps []time.Duration
	// MaximumInterval is the maximum number of days a Review-state card can be scheduled into the future.
	MaximumInterval int
	// EnableFuzzing determines whether to apply a small amount of random 'fuzz' to calculated intervals.
	EnableFuzzing bool
	// Internal fields
	decay  float64
	factor float64
}

func DefaultScheduler() *Scheduler {
	var params = DefaultParameters
	var decay = -params[20]

	if err := validateParameters(params); err != nil {
		panic(err)
	}

	return &Scheduler{
		Parameters:       params,
		DesiredRetention: 0.9,
		LearningSteps:    []time.Duration{1 * time.Minute, 10 * time.Minute},
		RelearningSteps:  []time.Duration{10 * time.Minute},
		MaximumInterval:  36500,
		EnableFuzzing:    true,
		decay:            decay,
		factor:           math.Pow(0.9, 1.0/decay) - 1,
	}
}

// validateParameters checks if the parameters are within valid bounds.
func validateParameters(parameters []float64) error {
	if len(parameters) != len(LowerBoundsParameters) {
		return fmt.Errorf("%w expected %d parameters, got %d", ErrInvalidParam, len(LowerBoundsParameters), len(parameters))
	}

	var errorMessages []string
	for i, param := range parameters {
		lowerBound := LowerBoundsParameters[i]
		upperBound := UpperBoundsParameters[i]
		if param < lowerBound || param > upperBound {
			errorMessages = append(errorMessages,
				fmt.Sprintf("parameters[%d] = %f is out of bounds: (%f, %f)", i, param, lowerBound, upperBound))
		}
	}

	if len(errorMessages) > 0 {
		return fmt.Errorf("%w one or more parameters are out of bounds:\n%s", ErrInvalidParam, strings.Join(errorMessages, "\n"))
	}

	return nil
}

func (s *Scheduler) GetCardRetrievability(card *Card, now time.Time) float64 {
	if card.LastReview == nil {
		return 0
	}

	// Calculate elapsed days
	elapsedDays := now.Sub(*card.LastReview).Hours() / 24
	stability := card.Stability

	// Calculate retrievability

	return math.Pow(1+s.factor*elapsedDays/stability, s.decay)
}

func (s *Scheduler) ReviewCard(card *Card, rating Rating, reviewDatetime time.Time, reviewDuration time.Duration) (*Card, *ReviewLog) {
	var (
		daysSinceLastReview float64
		hasLastReview       bool
		nextInterval        time.Duration
	)

	if card.LastReview != nil {
		hasLastReview = true
		daysSinceLastReview = reviewDatetime.Sub(*card.LastReview).Hours() / 24
	}

	// copy
	card = card.Duplicate()

	switch card.State {
	case Learning, Relearning:
		steps := s.LearningSteps
		if card.State == Relearning {
			steps = s.RelearningSteps
		}

		// update the card's stability and difficulty
		if card.Stability == 0 && card.Difficulty == 0 {
			card.Stability = s.initialStability(rating)
			card.Difficulty = s.initialDifficulty(rating)
		} else if hasLastReview && daysSinceLastReview < 1 {
			card.Stability = s.shortTermStability(card.Stability, rating)
			card.Difficulty = s.nextDifficulty(card.Difficulty, rating)
		} else {
			retrievability := s.GetCardRetrievability(card, reviewDatetime)
			card.Stability = s.nextStability(card.Difficulty, card.Stability, retrievability, rating)
			card.Difficulty = s.nextDifficulty(card.Difficulty, rating)
		}

		// update the card's next interval
		// Calculate next interval
		// Graduate card if it's past all steps or if there are no steps
		if len(steps) == 0 || (card.Step >= len(steps) && rating > Again) {
			card.State = Review
			card.Step = -1
			nextIntervalDays := s.nextInterval(card.Stability)
			nextInterval = time.Duration(nextIntervalDays) * 24 * time.Hour
		} else {
			switch rating {
			case Again:
				card.Step = 0
				nextInterval = steps[card.Step]
			case Hard:
				// Step does not change
				if card.Step == 0 {
					if len(steps) == 1 {
						nextInterval = time.Duration(float64(steps[0]) * 1.5)
					} else { // len >= 2
						nextInterval = (steps[0] + steps[1]) / 2
					}
				} else {
					nextInterval = steps[card.Step]
				}
			case Good:
				if card.Step+1 >= len(steps) { // Last step
					card.State = Review
					card.Step = -1
					nextIntervalDays := s.nextInterval(card.Stability)

					nextInterval = time.Duration(nextIntervalDays) * 24 * time.Hour
				} else {
					card.Step++
					nextInterval = steps[card.Step]
				}
			case Easy:
				card.State = Review
				card.Step = -1
				nextIntervalDays := s.nextInterval(card.Stability)
				nextInterval = time.Duration(nextIntervalDays) * 24 * time.Hour
			}
		}

	case Review:
		// Update stability and difficulty
		if hasLastReview && daysSinceLastReview < 1 {
			card.Stability = s.shortTermStability(card.Stability, rating)
		} else {
			retrievability := s.GetCardRetrievability(card, reviewDatetime)
			card.Stability = s.nextStability(card.Difficulty, card.Stability, retrievability, rating)
		}
		card.Difficulty = s.nextDifficulty(card.Difficulty, rating)

		// Calculate next interval
		switch rating {
		case Again:
			if len(s.RelearningSteps) == 0 {
				// Stay in Review state
				nextIntervalDays := s.nextInterval(card.Stability)
				nextInterval = time.Duration(nextIntervalDays) * 24 * time.Hour
			} else {
				// Enter Relearning state
				card.State = Relearning
				card.Step = 0
				nextInterval = s.RelearningSteps[card.Step]
			}
		default: // Hard, Good, Easy
			nextIntervalDays := s.nextInterval(card.Stability)
			nextInterval = time.Duration(nextIntervalDays) * 24 * time.Hour
		}

	default:
		panic(fmt.Sprintf("unknown state %v card id %v", card.ID, card.State))
	}

	if s.EnableFuzzing && card.State == Review {
		nextInterval = s.getFuzzedInterval(nextInterval)
	}

	// Finalize card update
	card.Due = reviewDatetime.Add(nextInterval)
	lastReviewTime := reviewDatetime
	card.LastReview = &lastReviewTime

	// Create review log
	reviewLog := &ReviewLog{
		CardID:         card.ID,
		Rating:         rating,
		ReviewDateTime: reviewDatetime,
	}
	if int64(reviewDuration) > 0 {
		durationSeconds := int64(reviewDuration.Seconds())
		reviewLog.ReviewDuration = &durationSeconds
	}

	return card, reviewLog
}

func (s *Scheduler) clampDdifficulty(difficulty float64) float64 {
	return min(max(difficulty, MinDifficulty), MaxDifficulty)
}

func (s *Scheduler) clampStability(stability float64) float64 {
	return max(stability, StabilityMin)
}

func (s *Scheduler) initialStability(rating Rating) float64 {
	return s.clampStability(s.Parameters[rating-1])
}

func (s *Scheduler) initialDifficulty(rating Rating) float64 {
	var (
		p4 = s.Parameters[4]
		p5 = s.Parameters[5]
	)

	difficulty := p4 - math.Pow(math.E, p5*(float64(rating)-1)) + 1

	return s.clampDdifficulty(difficulty)
}

func (s *Scheduler) nextInterval(stability float64) (days int) {
	decay := s.decay
	factor := s.factor

	nextInterval := (stability / factor) * (math.Pow(s.DesiredRetention, 1/decay) - 1)
	days = int(math.Round(nextInterval))

	// Ensure interval is at least 1 and not more than the maximum interval
	days = max(1, days)
	days = min(days, s.MaximumInterval)
	return days
}

func (s *Scheduler) shortTermStability(stability float64, rating Rating) float64 {
	p17 := s.Parameters[17]
	p18 := s.Parameters[18]
	p19 := s.Parameters[19]

	shortTermStabilityIncrease := math.Pow(math.E, p17*(float64(rating)-3+p18)) * math.Pow(stability, -p19)

	if rating == Good || rating == Easy {
		shortTermStabilityIncrease = max(shortTermStabilityIncrease, 1.0)
	}

	return s.clampStability(stability * shortTermStabilityIncrease)
}

func (s *Scheduler) nextDifficulty(difficulty float64, rating Rating) float64 {
	p6 := s.Parameters[6]
	p7 := s.Parameters[7]

	linearDamping := func(deltaDifficulty, difficulty float64) float64 {
		return (10.0 - difficulty) * deltaDifficulty / 9.0
	}

	meanReversion := func(arg1, arg2 float64) float64 {
		return p7*arg1 + (1-p7)*arg2
	}

	arg1 := s.initialDifficulty(Easy)
	deltaDifficulty := -(p6 * (float64(rating) - 3))
	arg2 := difficulty + linearDamping(deltaDifficulty, difficulty)

	return s.clampDdifficulty(meanReversion(arg1, arg2))
}

func (s *Scheduler) nextStability(difficulty, stability, retrievability float64, rating Rating) float64 {
	var nextStab float64

	if rating == Again {
		nextStab = s.nextForgetStability(difficulty, stability, retrievability)
	} else {
		// Handles Hard, Good, Easy
		nextStab = s.nextRecallStability(difficulty, stability, retrievability, rating)
	}

	return s.clampStability(nextStab)
}

func (s *Scheduler) nextForgetStability(difficulty, stability, retrievability float64) float64 {
	p11 := s.Parameters[11]
	p12 := s.Parameters[12]
	p13 := s.Parameters[13]
	p14 := s.Parameters[14]
	p17 := s.Parameters[17]
	p18 := s.Parameters[18]

	longTermParams := p11 *
		math.Pow(difficulty, -p12) *
		(math.Pow(stability+1, p13) - 1) *
		math.Pow(math.E, (1-retrievability)*p14)

	shortTermParams := stability / math.Pow(math.E, p17*p18)

	return min(longTermParams, shortTermParams)
}

func (s *Scheduler) nextRecallStability(difficulty, stability, retrievability float64, rating Rating) float64 {
	p8 := s.Parameters[8]
	p9 := s.Parameters[9]
	p10 := s.Parameters[10]
	p15 := s.Parameters[15]
	p16 := s.Parameters[16]

	hardPenalty := 1.0
	if rating == Hard {
		hardPenalty = p15
	}

	easyBonus := 1.0
	if rating == Easy {
		easyBonus = p16
	}

	return stability * (1 +
		math.Pow(math.E, p8)*
			(11-difficulty)*
			math.Pow(stability, -p9)*
			(math.Pow(math.E, (1-retrievability)*p10)-1)*
			hardPenalty*
			easyBonus)
}

func (s *Scheduler) getFuzzedInterval(interval time.Duration) time.Duration {
	intervalDays := float64(interval.Hours() / 24)

	if intervalDays < 2.5 {
		return interval
	}

	// Helper function to compute fuzz bounds
	getFuzzRange := func(days float64) (int, int) {
		delta := 1.0
		for _, fr := range FuzzRanges {
			delta += fr.Factor * max(0.0, min(days, fr.End)-fr.Start)
		}

		minIvl := int(math.Round(days - delta))
		maxIvl := int(math.Round(days + delta))

		minIvl = max(2, minIvl)
		maxIvl = min(maxIvl, s.MaximumInterval)
		minIvl = min(minIvl, maxIvl)

		return minIvl, maxIvl
	}

	minIvl, maxIvl := getFuzzRange(intervalDays)

	// Seed the random number generator
	rand.Seed(time.Now().UnixNano())

	fuzzedDays := float64(minIvl) + rand.Float64()*(float64(maxIvl-minIvl+1))
	fuzzedDays = math.Round(fuzzedDays)
	fuzzedDaysClamped := min(fuzzedDays, float64(s.MaximumInterval))

	return time.Duration(fuzzedDaysClamped) * 24 * time.Hour
}
