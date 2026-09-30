package tactical

import (
	"fmt"

	"github.com/iiooiioo888/qianjunpo-reborn/pkg/board"
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/combat"
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/validate"
)

// AoE error codes for stub skill wiring (submit-time validation).
const (
	CodeAoEOutOfBounds = "AOE_OUT_OF_BOUNDS"
	CodeAoEOutOfRange  = "AOE_OUT_OF_RANGE"
	CodeAoELoSBlocked  = "AOE_LOS_BLOCKED"
	CodeAoENoTargets   = "AOE_NO_TARGETS"
)

// AoEError is returned when an AoE command or ApplyAoEStrike precondition fails.
type AoEError struct {
	Code    string
	Message string
}

func (e AoEError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return e.Code
}

// CollectAoETargets is a thin match wrapper over combat.CollectAoETargets.
func (m *Match) CollectAoETargets(center board.Coord, radius int, excludeUnitID uint32) []uint32 {
	if m == nil || m.Board == nil {
		return nil
	}
	return combat.CollectAoETargets(m.Board, center, radius, excludeUnitID)
}

// enemyAoETargets returns hostile unit IDs in the splash, ascending by ID.
func (m *Match) enemyAoETargets(attacker *Unit, center board.Coord, radius int) []uint32 {
	ids := m.CollectAoETargets(center, radius, attacker.ID)
	if len(ids) == 0 {
		return nil
	}
	out := make([]uint32, 0, len(ids))
	for _, id := range ids {
		u := m.Units[id]
		if u == nil || u.Owner == attacker.Owner || u.Stats.HP.Raw() <= 0 {
			continue
		}
		out = append(out, id)
	}
	return out
}

func validateAoECenter(center board.Coord) error {
	if !board.InBounds(center) {
		return AoEError{Code: CodeAoEOutOfBounds, Message: "aoe center off board"}
	}
	return nil
}

// validateAoECast checks center bounds and the same range / LOS rules as single-target attacks.
func (m *Match) validateAoECast(attacker *Unit, center board.Coord) error {
	if err := validateAoECenter(center); err != nil {
		return err
	}
	if !combat.InAttackRange(attacker.Pos, center, attacker.Stats.Range) {
		return AoEError{Code: CodeAoEOutOfRange, Message: "aoe center out of attack range"}
	}
	centerUnitID := m.Board.GetUnit(center)
	if combat.RangedAttackNeedsLine(attacker.Stats.Range) &&
		!combat.AttackLineClear(m.Board, attacker.Pos, center, attacker.ID, centerUnitID) {
		return AoEError{Code: CodeAoELoSBlocked, Message: "aoe line of sight blocked"}
	}
	return nil
}

func (m *Match) validateAoE(cmd Command, attacker *Unit) error {
	if err := m.validateAoECast(attacker, cmd.To); err != nil {
		return err
	}
	if len(m.enemyAoETargets(attacker, cmd.To, combat.DefaultAoERadius)) == 0 {
		return AoEError{Code: CodeAoENoTargets, Message: "no enemy in aoe splash"}
	}
	return nil
}

// ApplyAoEStrike resolves strike damage for each enemy in CollectAoETargets(center, radius).
// The attacker is excluded from splash collection; friendlies in range are skipped.
func (m *Match) ApplyAoEStrike(attackerID uint32, center board.Coord, radius int) error {
	if m == nil {
		return fmt.Errorf("tactical: nil match")
	}
	attacker := m.Units[attackerID]
	if attacker == nil || attacker.Stats.HP.Raw() <= 0 {
		return validate.MoveError{Code: validate.CodeWrongStart, Message: "aoe attacker invalid"}
	}
	if err := m.validateAoECast(attacker, center); err != nil {
		return err
	}
	targets := m.enemyAoETargets(attacker, center, radius)
	if len(targets) == 0 {
		return AoEError{Code: CodeAoENoTargets, Message: "no enemy in aoe splash"}
	}
	for _, id := range targets {
		defender := m.Units[id]
		if defender == nil {
			continue
		}
		m.applyStrike(attacker, defender)
	}
	return nil
}
