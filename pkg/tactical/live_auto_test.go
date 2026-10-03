package tactical

import (
	"testing"
)

func TestLiveAutoOccupyBeforeWipeout(t *testing.T) {
	m := NewLiveMatchOccupy(0x0cc7a7)
	m.SetAutoCommandMode(AutoCommandBoth)
	m.RunSpectatorAuto(MaxTurnFrames + 8)
	if !m.Finished {
		t.Fatalf("expected finished, frame=%d", m.Frame)
	}
	if m.EndReason != EndCapture {
		t.Fatalf("want occupy (capture), got %s winner=%d frame=%d", EndReasonName(m.EndReason), m.Winner, m.Frame)
	}
	if m.Frame >= 43 {
		t.Fatalf("occupy should finish before typical wipeout frame (~43), frame=%d", m.Frame)
	}
	if m.Winner != 0 {
		t.Fatalf("expected P0 occupy win at spawn CP, winner=%d", m.Winner)
	}
}

func TestLiveAutoTimeoutReachesMaxTurnFrames(t *testing.T) {
	m := NewLiveMatchTimeout(0x710e07a)
	m.SetAutoCommandMode(AutoCommandBoth)
	m.RunSpectatorAuto(MaxTurnFrames + 8)
	if !m.Finished {
		t.Fatalf("expected finished, frame=%d", m.Frame)
	}
	if m.EndReason != EndTimeout {
		t.Fatalf("want timeout, got %s winner=%d frame=%d", EndReasonName(m.EndReason), m.Winner, m.Frame)
	}
	if m.Frame < MaxTurnFrames {
		t.Fatalf("expected to survive until timeout threshold, frame=%d want >=%d", m.Frame, MaxTurnFrames)
	}
}

func TestLiveControlPointsOccupyPreset(t *testing.T) {
	cps := LiveControlPointsOccupy()
	if len(cps) != 1 || cps[0].Pos != LiveSpawnP0 || cps[0].HoldFrames != LiveOccupyHoldFrames {
		t.Fatalf("unexpected occupy preset: %+v", cps)
	}
}
