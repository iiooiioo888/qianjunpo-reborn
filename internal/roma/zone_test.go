package roma

import (
	"testing"

	"github.com/iiooiioo888/qianjunpo-reborn/pkg/board"
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/tactical"
)

func TestZonePartitionInMemory(t *testing.T) {
	store := NewStore(nil)
	b, err := store.Join("cn-east-1", 0)
	if err != nil {
		t.Fatal(err)
	}
	if b.StateHash() == 0 {
		t.Fatal("expected non-zero hash")
	}
	_, hash, err := store.SubmitTacticalCommand(b.ID, 0, uint32(tactical.KindMove), tactical.UnitIDPlayer0, 5, 8, 0)
	if err != nil {
		t.Fatal(err)
	}
	frame, h2, _, _, _, err := store.StepLockstep(b.ID, 4, StepLockstepOpts{})
	if err != nil {
		t.Fatal(err)
	}
	if frame == 0 || h2 == 0 {
		t.Fatal("expected lockstep progress")
	}
	b2, _ := store.Get(b.ID)
	if hash != b2.StateHash() && h2 != b2.StateHash() {
		t.Fatalf("hash mismatch submit=%x step=%x state=%x", hash, h2, b2.StateHash())
	}
}

func TestJoinUsesLiveControlPoints(t *testing.T) {
	store := NewStore(nil)
	b, err := store.Join("capture-live", 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(liveJoinControlPoints) != 1 {
		t.Fatalf("points=%d", len(liveJoinControlPoints))
	}
	cp := liveJoinControlPoints[0]
	if cp.Pos != (board.Coord{X: 2, Y: 8}) || cp.HoldFrames != 3 {
		t.Fatalf("cp=%+v", cp)
	}
	for i := 0; i < 3 && !b.Match.Finished; i++ {
		if _, _, _, _, _, err := store.StepLockstep(b.ID, 1, StepLockstepOpts{}); err != nil {
			t.Fatal(err)
		}
	}
	if !b.Match.Finished || b.Match.EndReason != tactical.EndCapture || b.Match.Winner != 0 {
		t.Fatalf("finished=%v reason=%d winner=%d", b.Match.Finished, b.Match.EndReason, b.Match.Winner)
	}
}

func TestRedisMustNotStoreAuthoritativeFields(t *testing.T) {
	// Documentation test: authoritative fields live only on tactical.Match in memory.
	store := NewStore(nil)
	b, err := store.Join("doc", 1)
	if err != nil {
		t.Fatal(err)
	}
	u := b.Match.Units[tactical.UnitIDPlayer0]
	if u.Stats.HP.Raw() <= 0 {
		t.Fatal("hp authoritative in memory")
	}
}
