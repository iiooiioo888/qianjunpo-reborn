package roma

import (
	"errors"
	"fmt"

	"github.com/iiooiioo888/qianjunpo-reborn/pkg/board"
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/combat"
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/hash"
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/tactical"
)

const winnerUndecided = 255

func tacticalSeed(zoneID string, shard uint32) uint64 {
	h := hash.New()
	h.Write([]byte(zoneID))
	h.WriteUint64(uint64(shard))
	return h.Sum64()
}

func protoToCommand(playerID, kind, unitID uint32, toX, toY int32, skillID uint32) (tactical.Command, error) {
	if playerID >= tactical.PlayerCount {
		return tactical.Command{}, fmt.Errorf("roma: invalid player %d", playerID)
	}
	cmd := tactical.Command{
		PlayerID: uint8(playerID),
		Kind:     tactical.CommandKind(kind),
		UnitID:   unitID,
		To:       board.Coord{X: int(toX), Y: int(toY)},
	}
	switch cmd.Kind {
	case tactical.KindMove, tactical.KindAttack, tactical.KindPass:
		return cmd, nil
	case tactical.KindSkill:
		if skillID == 0 {
			return tactical.Command{}, fmt.Errorf("roma: kind skill requires skill_id")
		}
		cmd.SkillID = combat.SkillID(skillID)
		return cmd, nil
	default:
		return tactical.Command{}, fmt.Errorf("roma: unknown tactical kind %d", kind)
	}
}

func (s *Store) SubmitTacticalCommand(id BattleID, playerID, kind, unitID uint32, toX, toY int32, skillID uint32) (uint64, uint64, error) {
	cmd, err := protoToCommand(playerID, kind, unitID, toX, toY, skillID)
	if err != nil {
		return 0, 0, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	b, ok := s.battles[id]
	if !ok {
		return 0, 0, errors.New("roma: battle not found")
	}
	if b.Match == nil {
		return 0, 0, errors.New("roma: battle has no tactical match")
	}
	if err := b.Match.Submit(cmd); err != nil {
		return b.Match.Frame, b.Match.StateHash(), err
	}
	return b.Match.Frame, b.Match.StateHash(), nil
}

// StepLockstepOpts controls optional auto-command fill for spectator / auto-battle ticks.
type StepLockstepOpts struct {
	OneShotAuto      bool
	SetAutoModeWire  uint32 // 0=unchanged; see tactical.AutoCommandWire*
}

func (s *Store) StepLockstep(id BattleID, steps uint32, opts StepLockstepOpts) (uint64, uint64, bool, uint32, uint32, error) {
	if steps == 0 {
		steps = 1
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	b, ok := s.battles[id]
	if !ok {
		return 0, 0, false, winnerUndecided, 0, errors.New("roma: battle not found")
	}
	if b.Match == nil {
		return 0, 0, false, winnerUndecided, 0, errors.New("roma: battle has no tactical match")
	}
	if mode, apply := tactical.AutoCommandModeFromWire(opts.SetAutoModeWire); apply {
		b.Match.SetAutoCommandMode(mode)
	}
	for i := 0; i < int(steps); i++ {
		auto := opts.OneShotAuto || b.Match.AutoCommandMode() != tactical.AutoCommandOff
		if auto {
			b.Match.StepWithAuto()
		} else {
			b.Match.StepLockstep()
		}
		s.observeAndTickDilation(b)
	}
	winner := uint32(winnerUndecided)
	if b.Match.Finished {
		winner = uint32(b.Match.Winner)
	}
	autoWire := tactical.AutoCommandModeToWire(b.Match.AutoCommandMode())
	return b.Match.Frame, b.Match.StateHash(), b.Match.Finished, winner, autoWire, nil
}
