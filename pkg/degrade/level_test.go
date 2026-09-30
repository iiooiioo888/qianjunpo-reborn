package degrade

import "testing"

func TestEscalateMonotonic(t *testing.T) {
	m := NewMachine()
	for i := 0; i < 10; i++ {
		m.Escalate()
	}
	if m.Current != L5 {
		t.Fatalf("got %v", m.Current)
	}
}

func TestOrderedRecovery(t *testing.T) {
	m := NewMachine()
	m.EscalateTo(L5)
	path := []Level{m.Current}
	for m.CanRecover() {
		path = append(path, m.Recover())
	}
	want := []Level{L5, L4, L3, L2, L1, L0}
	if len(path) != len(want) {
		t.Fatalf("path=%v", path)
	}
	for i := range want {
		if path[i] != want[i] {
			t.Fatalf("step %d: got %v want %v", i, path[i], want[i])
		}
	}
}

func TestSyncSuggestedEscalatesImmediate(t *testing.T) {
	m := NewMachine()
	m.SyncSuggested(L3)
	if m.Current != L3 {
		t.Fatalf("got %v", m.Current)
	}
}

func TestLevelFromLoad(t *testing.T) {
	if LevelFromLoad(100, 10000) != L0 {
		t.Fatal("expected L0")
	}
	if LevelFromLoad(2000, 10000) != L5 {
		t.Fatal("expected L5")
	}
}
