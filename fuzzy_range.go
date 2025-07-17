package fsrs

import "math"

// FuzzRange represents a range for fuzzing intervals in FSRS.
type FuzzRange struct {
	Start  float64 `json:"start"`
	End    float64 `json:"end"`
	Factor float64 `json:"factor"`
}

// FuzzRanges represents the fuzzing ranges for FSRS scheduling.
var FuzzRanges = []FuzzRange{
	{
		Start:  2.5,
		End:    7.0,
		Factor: 0.15,
	},
	{
		Start:  7.0,
		End:    20.0,
		Factor: 0.1,
	},
	{
		Start:  20.0,
		End:    math.Inf(1), // Positive infinity
		Factor: 0.05,
	},
}
