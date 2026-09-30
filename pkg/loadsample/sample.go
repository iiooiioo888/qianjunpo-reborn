// Package loadsample provides trigger sampling hooks for the time dilation control loop.
package loadsample

// Sample is one observation used to drive dilation and degradation.
type Sample struct {
	CPUPercent    float64 // 0..100
	QueueLength   int
	GrowthRate    float64 // commands/sec derivative
	P99LatencyMs  float64
}

// Predictor is a pluggable load forecaster (LSTM or other ML backend).
type Predictor interface {
	Predict(next Sample) (predictedQueue int, predictedP99Ms float64)
}

// NoOpPredictor returns the current sample unchanged.
type NoOpPredictor struct{}

func (NoOpPredictor) Predict(s Sample) (int, float64) {
	return s.QueueLength, s.P99LatencyMs
}
