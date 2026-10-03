package tactical

import (
	"testing"
)

func TestIdleStepLockstepAdvancesFrame(t *testing.T) {
	m := NewMatch(42)
	start := m.Frame
	for i := 0; i < 5; i++ {
		m.StepLockstep()
	}
	if m.Frame != start+5 {
		t.Fatalf("frame=%d want %d", m.Frame, start+5)
	}
}

func TestSpectatorAutoNoManualInputFinishes(t *testing.T) {
	m := NewMatch(0xa110)
	start := m.Frame
	m.RunSpectatorAuto(maxTurnFrames + 8)
	if m.Frame <= start {
		t.Fatalf("expected frames to advance, frame=%d", m.Frame)
	}
	if !m.Finished {
		t.Fatalf("expected finished match, frame=%d", m.Frame)
	}
	switch m.EndReason {
	case EndAnnihilation, EndTimeout, EndMutualWipe:
	default:
		t.Fatalf("unexpected end reason %d (%s)", m.EndReason, EndReasonName(m.EndReason))
	}
}

func TestSpectatorAutoCanReachWipeout(t *testing.T) {
	m := NewMatch(0xdeadbeef)
	m.RunSpectatorAuto(maxTurnFrames + 8)
	if !m.Finished {
		t.Fatal("match did not finish")
	}
	if m.EndReason != EndAnnihilation {
		t.Fatalf("want wipeout (annihilation), got %s winner=%d frame=%d", EndReasonName(m.EndReason), m.Winner, m.Frame)
	}
	if m.Winner > 1 {
		t.Fatalf("invalid winner %d", m.Winner)
	}
	if EndReasonName(m.EndReason) != "wipeout" {
		t.Fatalf("Live token mismatch: %q", EndReasonName(m.EndReason))
	}
}

func TestAutoMissingOnlyFillsAbsentPlayer(t *testing.T) {
	m := NewMatch(7)
	m.SetAutoCommandMode(AutoCommandMissing)
	u0 := m.Units[UnitIDPlayer0].Pos
	if err := m.Submit(Command{PlayerID: 0, Kind: KindPass, UnitID: UnitIDPlayer0, To: u0}); err != nil {
		t.Fatal(err)
	}
	m.StepWithAuto()
	if m.PendingCommandCount() < 2 {
		t.Fatalf("expected P0 pass + P1 auto queued, pending=%d", m.PendingCommandCount())
	}
	before := m.Units[UnitIDPlayer1].Pos
	for !m.Finished && m.Frame < 6 {
		m.SetAutoCommandMode(AutoCommandMissing)
		_ = m.Submit(Command{PlayerID: 0, Kind: KindPass, UnitID: UnitIDPlayer0, To: m.Units[UnitIDPlayer0].Pos})
		m.StepWithAuto()
	}
	if m.Units[UnitIDPlayer1].Pos == before {
		t.Fatal("expected absent player P1 auto commands to eventually move")
	}
}
