// Package anticheat provides a skeleton AI anti-cheat pipeline: feature extraction
// and pluggable model inference (no real model download in CI).
package anticheat

import (
	"context"
	"errors"
)

// FeatureVector is a fixed-size gameplay signal snapshot for inference.
type FeatureVector struct {
	PlayerID     uint64
	SpeedPPM     int32 // parts-per-million of max allowed speed
	AimDeltaPPM  int32
	CommandGapMs int32
}

// Extractor builds feature vectors from raw telemetry (stub).
type Extractor interface {
	Extract(ctx context.Context, raw map[string]int64) (FeatureVector, error)
}

// StubExtractor maps known keys for tests and demos.
type StubExtractor struct{}

func (StubExtractor) Extract(_ context.Context, raw map[string]int64) (FeatureVector, error) {
	if len(raw) == 0 {
		return FeatureVector{}, errors.New("anticheat: empty telemetry")
	}
	return FeatureVector{
		PlayerID:     uint64(raw["player_id"]),
		SpeedPPM:     int32(raw["speed_ppm"]),
		AimDeltaPPM:  int32(raw["aim_ppm"]),
		CommandGapMs: int32(raw["gap_ms"]),
	}, nil
}

// Verdict is the model output.
type Verdict struct {
	CheatScore float64 // 0..1
	Label      string
}

// Model performs inference on feature vectors.
type Model interface {
	Infer(ctx context.Context, fv FeatureVector) (Verdict, error)
}

// FakeModel returns deterministic scores from feature thresholds.
type FakeModel struct {
	ThresholdPPM int32
}

func (m FakeModel) Infer(_ context.Context, fv FeatureVector) (Verdict, error) {
	th := m.ThresholdPPM
	if th == 0 {
		th = 1_050_000
	}
	score := 0.0
	label := "clean"
	if fv.SpeedPPM > th || fv.AimDeltaPPM > th {
		score = 0.92
		label = "suspicious"
	}
	return Verdict{CheatScore: score, Label: label}, nil
}

// Pipeline runs extract → infer end-to-end.
type Pipeline struct {
	Extractor Extractor
	Model     Model
}

func (p Pipeline) Evaluate(ctx context.Context, raw map[string]int64) (Verdict, error) {
	if p.Extractor == nil || p.Model == nil {
		return Verdict{}, errors.New("anticheat: pipeline not configured")
	}
	fv, err := p.Extractor.Extract(ctx, raw)
	if err != nil {
		return Verdict{}, err
	}
	return p.Model.Infer(ctx, fv)
}
