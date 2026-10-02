package integration

import (
	"testing"

	"github.com/iiooiioo888/qianjunpo-reborn/pkg/board"
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/tactical"
)

// Cross-package smoke: default duel replay unchanged; objective duels use explicit control points.
func TestPhase2VictoryConditionsIntegration(t *testing.T) {
	TestPhase2TacticalVerticalSlice(t)

	mCap := tactical.NewMatchWithControlPoints(7, []tactical.ControlPoint{
		{Pos: board.Coord{2, 8}, HoldFrames: 2},
	})
	for i := 0; i < 2 && !mCap.Finished; i++ {
		mCap.StepLockstep()
	}
	if mCap.EndReason != tactical.EndCapture {
		t.Fatalf("capture path: reason=%d", mCap.EndReason)
	}

	mAnn := tactical.NewMatch(8)
	def := mAnn.Units[tactical.UnitIDPlayer1]
	def.Stats.HP = def.Stats.HP.Sub(def.Stats.HP)
	mAnn.Board.ClearUnit(def.Pos)
	mAnn.StepLockstep()
	if mAnn.EndReason != tactical.EndAnnihilation {
		t.Fatalf("annihilation path: reason=%d", mAnn.EndReason)
	}
}
