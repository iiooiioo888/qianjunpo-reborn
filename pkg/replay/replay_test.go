package replay

import (
	"testing"

	"github.com/iiooiioo888/qianjunpo-reborn/pkg/rng"
)

func TestRoundTrip(t *testing.T) {
	seed := rng.State{S0: 1, S1: 2}
	rec := NewRecorder(0xABC, seed)
	rec.AddFrame(FrameCommand{Frame: 1, Player: 0, Payload: []byte("move")})
	done := rec.Finish(0xDEF)
	if err := VerifyConsistency(done); err != nil {
		t.Fatal(err)
	}
	raw := Marshal(done)
	back, err := Unmarshal(raw)
	if err != nil {
		t.Fatal(err)
	}
	if back.FinalHash != done.FinalHash || len(back.Frames) != 1 {
		t.Fatal("round trip")
	}
}

func TestGzipRoundTrip(t *testing.T) {
	seed := rng.State{S0: 9, S1: 8}
	rec := NewRecorder(1, seed)
	rec.AddFrame(FrameCommand{Frame: 0, Player: 1, Payload: []byte{1, 2, 3}})
	done := rec.Finish(99)
	z, err := MarshalGzip(done)
	if err != nil {
		t.Fatal(err)
	}
	back, err := UnmarshalGzip(z)
	if err != nil {
		t.Fatal(err)
	}
	if err := VerifyConsistency(back); err != nil {
		t.Fatal(err)
	}
}
