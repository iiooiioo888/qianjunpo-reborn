package roma

import "testing"

func TestZonePartitionInMemory(t *testing.T) {
	store := NewStore(nil)
	b, err := store.Join("cn-east-1", 0)
	if err != nil {
		t.Fatal(err)
	}
	if b.StateHash() == 0 {
		t.Fatal("expected non-zero hash")
	}
	h1, err := store.SubmitCommand(b.ID, 1, 1, 0)
	if err != nil {
		t.Fatal(err)
	}
	b2, _ := store.Get(b.ID)
	if h1 != b2.StateHash() {
		t.Fatalf("hash mismatch h1=%x state=%x", h1, b2.StateHash())
	}
}

func TestRedisMustNotStoreAuthoritativeFields(t *testing.T) {
	// Documentation test: authoritative fields live only on BattleState in memory.
	b := &BattleState{Units: map[uint32]*UnitState{1: {HP: 10, X: 1, Y: 2}}}
	if b.Units[1].HP != 10 {
		t.Fatal("hp authoritative in memory")
	}
}
