package loadpredict

import (
	"testing"
	"time"

	"github.com/iiooiioo888/qianjunpo-reborn/pkg/loadsample"
)

func TestSeriesPredictorFakeSeries(t *testing.T) {
	series := make([]loadsample.Sample, 400)
	for i := range series {
		series[i] = loadsample.Sample{
			QueueLength:  980 + i,
			CPUPercent:   98,
			P99LatencyMs: 500,
		}
	}
	pred := NewSeriesPredictor(series)
	cur := loadsample.Sample{QueueLength: 100, CPUPercent: 40, P99LatencyMs: 50}
	q, p99 := pred.Predict(cur)
	if q < 850 {
		t.Fatalf("expected 30s-ahead queue ≥850 from ramp series, got %d", q)
	}
	if p99 < 400 {
		t.Fatalf("expected 30s-ahead p99 ≥400, got %v", p99)
	}
	load := pred.ForecastLoad(cur)
	if !HookHighLoad(load) {
		t.Fatalf("expected high load hook at end of ramp, load=%v", load)
	}
}

func TestPreScalePlannerLeadWindow(t *testing.T) {
	planner := DefaultPreScalePlanner()
	now := time.Unix(1_700_000_000, 0)
	sig, ok := planner.Plan(now, 0.92, 850)
	if !ok {
		t.Fatal("expected pre-scale signal")
	}
	lead := sig.EffectiveAt.Sub(sig.EmittedAt)
	if lead < PreScaleLeadMin || lead > PreScaleLeadMax {
		t.Fatalf("lead %v outside [%v,%v]", lead, PreScaleLeadMin, PreScaleLeadMax)
	}
	if sig.TargetBuffer < 2 {
		t.Fatalf("expected buffer target ≥2, got %d", sig.TargetBuffer)
	}
	if !sig.RecoverTimeFlow {
		t.Fatal("expected timedilation recover hint")
	}
}

func TestHookHighLoadNegative(t *testing.T) {
	s := loadsample.Sample{QueueLength: 50, CPUPercent: 20, P99LatencyMs: 30}
	if HookHighLoad(CompositeLoad(s)) {
		t.Fatal("low sample should not trigger hook")
	}
}
