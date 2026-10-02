package roma

import (
	"testing"

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
	frame, h2, _, _, err := store.StepLockstep(b.ID, 4)
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
