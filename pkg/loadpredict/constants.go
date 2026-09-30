package loadpredict

import "time"

const (
	// ForecastHorizon is how far ahead the stub model looks (30 seconds).
	ForecastHorizon = 30 * time.Second
	// HighLoadFraction triggers pre-scale when predicted composite load ≥ this value.
	HighLoadFraction = 0.90
	// PreScaleLeadMin is the earliest lead time for Agones scale-up signals.
	PreScaleLeadMin = 7 * time.Minute
	// PreScaleLeadMax is the latest lead time for Agones scale-up signals.
	PreScaleLeadMax = 15 * time.Minute
	// SampleInterval matches the time-dilation control loop default (10 Hz).
	SampleInterval = 100 * time.Millisecond
)
