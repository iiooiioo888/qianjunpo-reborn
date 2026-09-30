package integration

import (
	"context"
	"testing"

	"github.com/iiooiioo888/qianjunpo-reborn/pkg/ai"
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/sim"
)

func TestPhase3AISnapshotToCommand(t *testing.T) {
	eng := sim.NewEngine(7)
	snap := sim.TakeSnapshot(eng.State, eng.RNG)
	pipe := ai.Pipeline{
		Planner:  &ai.MockPlanner{Cadence: 0},
		Tactical: ai.NewTacticalExecutor(0),
	}
	cmd, ok := pipe.RunStep(context.Background(), snap)
	if !ok {
		t.Fatal("expected command")
	}
	if cmd.MoveX == 0 && cmd.MoveY == 0 {
		t.Fatalf("cmd=%+v", cmd)
	}
}

func TestInferFallbackWhenDown(t *testing.T) {
	c := &ai.InferClient{BaseURL: "http://127.0.0.1:1"}
	out := c.Infer(context.Background(), "guard", "hold")
	if out.Text != ai.NPCFallback("guard") {
		t.Fatalf("got %q", out.Text)
	}
}
