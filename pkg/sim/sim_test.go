package sim

import (
	"testing"

	"github.com/iiooiioo888/qianjunpo-reborn/pkg/lockstep"
)

func demoSchedule(frames int) []ScheduledCommand {
	var out []ScheduledCommand
	for f := uint64(0); f < uint64(frames); f++ {
		out = append(out, ScheduledCommand{
			SubmitFrame: f,
			Cmd: lockstep.CommandPacket{
				PlayerID: uint8(f % 2),
				MoveX:    int8((f % 3) + 1),
				MoveY:    int8((f % 2)),
			},
		})
	}
	return out
}

func TestTwoClientsDeterministic(t *testing.T) {
	const frames = 40
	sched := demoSchedule(frames)
	a := NewEngine(0xdeadbeef)
	b := NewEngine(0xdeadbeef)
	ha := a.RunWithSchedule(sched, uint64(frames))
	hb := b.RunWithSchedule(sched, uint64(frames))
	if ha != hb {
		t.Fatalf("desync: %x vs %x", ha, hb)
	}
	for i := range a.History {
		if a.History[i] != b.History[i] {
			t.Fatalf("history diverged at %d", i)
		}
	}
}

func TestOptimisticEmptyMatches(t *testing.T) {
	const frames = 20
	sched := demoSchedule(frames)
	full := NewEngine(42)
	hFull := full.RunWithSchedule(sched, uint64(frames))

	// Client missing every other submission uses optimistic empty — both sides must agree.
	sparse := make([]ScheduledCommand, 0)
	for i, sc := range sched {
		if i%2 == 0 {
			sparse = append(sparse, sc)
		}
	}
	mixedA := NewEngine(42)
	mixedB := NewEngine(42)
	ha := mixedA.RunWithSchedule(sparse, uint64(frames))
	hb := mixedB.RunWithSchedule(sparse, uint64(frames))
	if ha != hb {
		t.Fatal("sparse clients desynced each other")
	}
	// Sparse schedule differs from full schedule; hashes should differ from full run.
	if ha == hFull {
		t.Fatal("expected sparse schedule to change outcome vs full")
	}
}

func TestRollbackResimulate(t *testing.T) {
	const target = 50
	const snapAt = 30
	sched := demoSchedule(target)

	linear := NewEngine(7)
	hLinear := linear.RunWithSchedule(sched, uint64(target))

	rewind := NewEngine(7)
	_ = rewind.RunWithSchedule(sched, snapAt)
	snap := TakeSnapshot(rewind.State, rewind.RNG)

	rollback := NewEngine(7)
	hRollback := rollback.ResimulateFromSnapshot(snap, sched, uint64(target))
	if hRollback != hLinear {
		t.Fatalf("rollback resim mismatch: %x vs %x", hRollback, hLinear)
	}
}

func TestHashPerformanceBudget(t *testing.T) {
	gs := NewGameState()
	r := NewEngine(1).RNG
	start := testing.Benchmark(func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			Hash(gs, r.GetState())
		}
	})
	// Per-iteration ns; target < 1ms (1e6 ns) per frame — smoke check only.
	if start.NsPerOp() > 1_000_000 {
		t.Fatalf("hash too slow: %d ns/op", start.NsPerOp())
	}
}
