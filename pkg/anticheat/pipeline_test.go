package anticheat

import (
	"context"
	"testing"
)

func TestPipelineEndToEndFakeData(t *testing.T) {
	p := Pipeline{Extractor: StubExtractor{}, Model: FakeModel{}}
	v, err := p.Evaluate(context.Background(), map[string]int64{
		"player_id": 42,
		"speed_ppm": 2_000_000,
		"aim_ppm":   100,
		"gap_ms":    50,
	})
	if err != nil {
		t.Fatal(err)
	}
	if v.Label != "suspicious" || v.CheatScore < 0.9 {
		t.Fatalf("verdict=%+v", v)
	}
}

func TestPipelineCleanPlayer(t *testing.T) {
	p := Pipeline{Extractor: StubExtractor{}, Model: FakeModel{}}
	v, err := p.Evaluate(context.Background(), map[string]int64{
		"player_id": 1,
		"speed_ppm": 500_000,
		"aim_ppm":   200_000,
		"gap_ms":    100,
	})
	if err != nil {
		t.Fatal(err)
	}
	if v.Label != "clean" {
		t.Fatalf("verdict=%+v", v)
	}
}
