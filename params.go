package fsrs

// Constants for FSRS parameters
const (
	StabilityMin        = 0.001
	InitialStabilityMax = 100.0

	MinDifficulty = 1.0
	MaxDifficulty = 10.0
)

// DefaultParameters represents the default FSRS parameters.
var DefaultParameters = []float64{
	0.2172,
	1.1771,
	3.2602,
	16.1507,
	7.0114,
	0.57,
	2.0966,
	0.0069,
	1.5261,
	0.112,
	1.0178,
	1.849,
	0.1133,
	0.3127,
	2.2934,
	0.2191,
	3.0004,
	0.7536,
	0.3332,
	0.1437,
	0.2,
}

// LowerBoundsParameters represents the lower bounds for FSRS parameters.
var LowerBoundsParameters = []float64{
	StabilityMin,
	StabilityMin,
	StabilityMin,
	StabilityMin,
	1.0,
	0.001,
	0.001,
	0.001,
	0.0,
	0.0,
	0.001,
	0.001,
	0.001,
	0.001,
	0.0,
	0.0,
	1.0,
	0.0,
	0.0,
	0.0,
	0.1,
}

// UpperBoundsParameters represents the upper bounds for FSRS parameters.
var UpperBoundsParameters = []float64{
	InitialStabilityMax,
	InitialStabilityMax,
	InitialStabilityMax,
	InitialStabilityMax,
	10.0,
	4.0,
	4.0,
	0.75,
	4.5,
	0.8,
	3.5,
	5.0,
	0.25,
	0.9,
	4.0,
	1.0,
	6.0,
	2.0,
	2.0,
	0.8,
	0.8,
}
