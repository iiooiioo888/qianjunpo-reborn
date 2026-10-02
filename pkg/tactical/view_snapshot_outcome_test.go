package tactical

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/iiooiioo888/qianjunpo-reborn/pkg/fixed"
)

func TestMatchToViewSnapshotOmitsOutcomeWhileInProgress(t *testing.T) {
	m := NewMatch(99)
	snap := MatchToViewSnapshot(m, 10000)
	if snap.Finished || snap.Winner != nil || snap.EndReason != EndNone {
		t.Fatalf("expected no outcome fields, got finished=%v winner=%v reason=%s", snap.Finished, snap.Winner, snap.EndReason)
	}
	raw, err := json.Marshal(snap)
	if err != nil {
		t.Fatal(err)
	}
	body := string(raw)
	if strings.Contains(body, `"finished"`) || strings.Contains(body, `"winner"`) || strings.Contains(body, `"endReason"`) {
		t.Fatalf("in-progress snapshot should omit outcome keys: %s", raw)
	}
}

func TestMatchToViewSnapshotAnnihilationOutcomeJSON(t *testing.T) {
	m := NewMatch(42)
	def := m.Units[UnitIDPlayer1]
	def.Stats.HP = fixed.Zero
	m.Board.ClearUnit(def.Pos)
	m.checkVictory()

	snap := MatchToViewSnapshot(m, 10000)
	if !snap.Finished || snap.Winner == nil || *snap.Winner != 0 || snap.EndReason != EndAnnihilation {
		t.Fatalf("snap outcome: finished=%v winner=%v reason=%s", snap.Finished, snap.Winner, snap.EndReason)
	}
	raw, err := json.Marshal(snap)
	if err != nil {
		t.Fatal(err)
	}
	var generic map[string]json.RawMessage
	if err := json.Unmarshal(raw, &generic); err != nil {
		t.Fatal(err)
	}
	if generic["finished"] == nil || generic["winner"] == nil || generic["endReason"] == nil {
		t.Fatalf("missing outcome keys in %s", raw)
	}
	var reason string
	if err := json.Unmarshal(generic["endReason"], &reason); err != nil {
		t.Fatal(err)
	}
	if reason != "EndAnnihilation" {
		t.Fatalf("endReason=%q", reason)
	}
}
