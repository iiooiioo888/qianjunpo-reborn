package integration

import (
	"testing"

	"github.com/iiooiioo888/qianjunpo-reborn/pkg/board"
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/hash"
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/replay"
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/rng"
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/validate"
)

func TestMovePipelineAndReplay(t *testing.T) {
	b := board.New()
	from := board.Coord{5, 5}
	to := board.Coord{8, 7}
	unitID := uint32(100)
	b.SetUnit(from, unitID)

	v := validate.NewValidator(b)
	req := validate.MoveRequest{
		UnitID: unitID, From: from, To: to, MovePoints: 10,
	}
	path, err := v.ValidateMove(req)
	if err != nil {
		t.Fatalf("validate: %v", err)
	}
	if err := v.ApplyMove(req, path); err != nil {
		t.Fatalf("apply: %v", err)
	}
	if b.GetUnit(to) != unitID || b.GetUnit(from) != 0 {
		t.Fatal("board update")
	}

	h := hash.New()
	b.Hash(func(v uint64) { h.WriteUint64(v) })
	stateHash := h.Sum64()

	seed := rng.State{S0: 12345, S1: 67890}
	rec := replay.NewRecorder(stateHash, seed)
	rec.AddFrame(replay.FrameCommand{
		Frame: 1, Player: 0, Payload: []byte{byte(from.X), byte(from.Y), byte(to.X), byte(to.Y)},
	})
	done := rec.Finish(stateHash)
	if err := replay.VerifyConsistency(done); err != nil {
		t.Fatal(err)
	}
	gz, err := replay.MarshalGzip(done)
	if err != nil {
		t.Fatal(err)
	}
	back, err := replay.UnmarshalGzip(gz)
	if err != nil {
		t.Fatal(err)
	}
	if back.InitialHash != stateHash {
		t.Fatal("hash preserved")
	}
}
