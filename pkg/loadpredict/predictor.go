package loadpredict

import (
	"time"

	"github.com/iiooiioo888/qianjunpo-reborn/pkg/loadsample"
)

// PreScaleSignal is emitted toward Agones buffer autoscaler / Fleet hooks (skeleton).
type PreScaleSignal struct {
	EmittedAt       time.Time
	EffectiveAt     time.Time
	TargetBuffer    int
	PredictedLoad   float64
	Horizon         time.Duration
	Reason          string
	RecoverTimeFlow bool // hint timedilation to BeginCatchUp after scale settles
}

// PreScalePlanner turns high-load forecasts into lead-time scale hints.
type PreScalePlanner struct {
	LeadMin time.Duration
	LeadMax time.Duration
}

// DefaultPreScalePlanner uses whitepaper 7–15 minute lead windows.
func DefaultPreScalePlanner() PreScalePlanner {
	return PreScalePlanner{LeadMin: PreScaleLeadMin, LeadMax: PreScaleLeadMax}
}

// Plan returns a signal when predictedLoad ≥ HighLoadFraction; ok is false otherwise.
func (p PreScalePlanner) Plan(now time.Time, predictedLoad float64, queueLen int) (PreScaleSignal, bool) {
	if predictedLoad < HighLoadFraction {
		return PreScaleSignal{}, false
	}
	lead := p.LeadMin + (p.LeadMax-p.LeadMin)/2
	buffer := 1 + queueLen/200
	if buffer < 2 {
		buffer = 2
	}
	return PreScaleSignal{
		EmittedAt:       now,
		EffectiveAt:     now.Add(lead),
		TargetBuffer:    buffer,
		PredictedLoad:   predictedLoad,
		Horizon:         ForecastHorizon,
		Reason:          "predicted_load_ge_90pct_30s",
		RecoverTimeFlow: true,
	}, true
}

// SeriesPredictor implements loadsample.Predictor using a deterministic fake series (CI-friendly).
type SeriesPredictor struct {
	Series []loadsample.Sample
	cursor int
}

// NewSeriesPredictor wraps a fixed time series for unit tests and offline replay.
func NewSeriesPredictor(series []loadsample.Sample) *SeriesPredictor {
	return &SeriesPredictor{Series: series}
}

func (p *SeriesPredictor) futureSample(current loadsample.Sample) loadsample.Sample {
	if len(p.Series) == 0 {
		return current
	}
	steps := int(ForecastHorizon / SampleInterval)
	if steps < 1 {
		steps = 1
	}
	idx := p.cursor + steps
	if idx >= len(p.Series) {
		idx = len(p.Series) - 1
	}
	return p.Series[idx]
}

// Predict advances the series cursor and returns values ForecastHorizon ahead (stub LSTM).
func (p *SeriesPredictor) Predict(_ loadsample.Sample) (predictedQueue int, predictedP99Ms float64) {
	f := p.futureSample(loadsample.Sample{})
	p.cursor++
	if p.cursor >= len(p.Series) {
		p.cursor = len(p.Series) - 1
	}
	return f.QueueLength, f.P99LatencyMs
}

// ForecastLoad returns composite load at the 30s-ahead point in the series.
func (p *SeriesPredictor) ForecastLoad(current loadsample.Sample) float64 {
	return CompositeLoad(p.futureSample(current))
}

// HookHighLoad reports whether the 30s-ahead forecast crosses the ≥90% threshold.
func HookHighLoad(forecastLoad float64) bool {
	return forecastLoad >= HighLoadFraction
}
