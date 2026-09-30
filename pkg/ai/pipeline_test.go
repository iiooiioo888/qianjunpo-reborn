package ai

import (
	"context"
	"testing"
	"time"

	"github.com/iiooiioo888/qianjunpo-reborn/pkg/sim"
)

func TestPipelineSnapshotToCommand(t *testing.T) {
	eng := sim.NewEngine(42)
	snap := sim.TakeSnapshot(eng.State, eng.RNG)
	planner := &MockPlanner{Cadence: 0}
	tactical := NewTacticalExecutor(1)
	pipe := Pipeline{Planner: planner, Tactical: tactical}
	cmd, ok := pipe.RunStep(context.Background(), snap)
	if !ok {
		t.Fatal("expected tactical command")
	}
	if cmd.PlayerID != 1 || cmd.MoveX == 0 && cmd.MoveY == 0 {
		t.Fatalf("cmd=%+v", cmd)
	}
}

func TestNPCFallbackFast(t *testing.T) {
	start := time.Now()
	_ = NPCFallback("guard")
	if time.Since(start) > time.Second {
		t.Fatal("fallback should be immediate")
	}
}
