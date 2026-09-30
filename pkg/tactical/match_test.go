package tactical

import (
	"testing"

	"github.com/iiooiioo888/qianjunpo-reborn/pkg/board"
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/replay"
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/validate"
)

func TestDemoMatchReplayRoundTrip(t *testing.T) {
	sched := DemoSchedule()
	m := NewMatch(0xcafe)
	target := DemoTargetFrame()
	final := m.RunSchedule(sched, target)
	rec := m.Recording()

	if err := replay.VerifyConsistency(rec); err != nil {
		t.Fatal(err)
	}
	got, err := ReplayFromRecording(rec)
	if err != nil {
		t.Fatal(err)
	}
	if got != final {
		t.Fatalf("replay hash %016x != recorded %016x", got, final)
	}
}

func TestIllegalMoveRejected(t *testing.T) {
	m := NewMatch(1)
	err := m.Submit(Command{
		PlayerID: 0,
		Kind:     KindMove,
		UnitID:   UnitIDPlayer0,
		To:       board.Coord{18, 18},
	})
	if err == nil {
		t.Fatal("expected error")
	}
	var me validate.MoveError
	if ok := asMoveError(err, &me); !ok || me.Code != validate.CodeExceedsMove {
		t.Fatalf("expected EXCEEDS_MOVE, got %v", err)
	}
}

func TestRiverBlockedWithoutBridge(t *testing.T) {
	m := NewMatch(2)
	err := m.Submit(Command{
		PlayerID: 0,
		Kind:     KindMove,
		UnitID:   UnitIDPlayer0,
		To:       board.Coord{5, 9},
	})
	if err == nil {
		t.Fatal("expected blocked river")
	}
}

func TestTwoClientsDeterministic(t *testing.T) {
	sched := DemoSchedule()
	target := DemoTargetFrame()
	a := NewMatch(99)
	b := NewMatch(99)
	ha := a.RunSchedule(sched, target)
	hb := b.RunSchedule(sched, target)
	if ha != hb {
		t.Fatalf("desync %016x vs %016x", ha, hb)
	}
}

func asMoveError(err error, out *validate.MoveError) bool {
	if err == nil {
		return false
	}
	if me, ok := err.(validate.MoveError); ok {
		*out = me
		return true
	}
	return false
}
