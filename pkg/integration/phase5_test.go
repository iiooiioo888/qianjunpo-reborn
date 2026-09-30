package integration

import (
	"testing"
	"time"

	"github.com/iiooiioo888/qianjunpo-reborn/pkg/loadpredict"
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/loadsample"
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/timedilation"
)

func TestPhase5PredictScaleDilationChain(t *testing.T) {
	series := make([]loadsample.Sample, 500)
	for i := range series {
		series[i] = loadsample.Sample{QueueLength: 980 + i, CPUPercent: 98, P99LatencyMs: 500}
	}
	pred := loadpredict.NewSeriesPredictor(series)
	clock := timedilation.NewManualClock(time.Unix(0, 0))
	ctrl := timedilation.NewController(timedilation.DefaultControllerConfig(), clock, pred)

	sample := loadsample.Sample{QueueLength: 100, CPUPercent: 70, P99LatencyMs: 200}
	load := pred.ForecastLoad(sample)
	if !loadpredict.HookHighLoad(load) {
		t.Fatal("expected high load for integration chain")
	}
	sig, ok := loadpredict.DefaultPreScalePlanner().Plan(clock.Now(), load, sample.QueueLength)
	if !ok || !sig.RecoverTimeFlow {
		t.Fatalf("prescale signal missing: ok=%v sig=%+v", ok, sig)
	}

	rec := ctrl.Step(sample, 1)
	if rec.Rate >= timedilation.MaxRate {
		t.Fatalf("predictor should pull rate down under hot forecast, rate=%v", rec.Rate)
	}
}
