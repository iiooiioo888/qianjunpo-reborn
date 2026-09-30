package replay

import (
	"testing"

	"github.com/iiooiioo888/qianjunpo-reborn/pkg/rng"
)

func TestVerifyTerminal(t *testing.T) {
	seed := rng.State{S0: 1, S1: 2}
	rec := NewRecorder(10, seed)
	rec.AddFrame(FrameCommand{Frame: 0, Player: 0, Payload: []byte{1}})
	done := rec.Finish(42)
	if err := VerifyTerminal(done, 42); err != nil {
		t.Fatal(err)
	}
	if err := VerifyTerminal(done, 43); err == nil {
		t.Fatal("expected mismatch")
	}
}
