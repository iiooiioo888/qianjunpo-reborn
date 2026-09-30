// Package combat provides deterministic unit stats, counter matrix, and damage helpers.
// Rules are loaded from JSON (see docs/combat-formula.md); simulation uses FP64 only.
package combat

import "github.com/iiooiioo888/qianjunpo-reborn/pkg/fixed"

// UnitType is infantry / archer / cavalry.
type UnitType uint8

const (
	UnitInfantry UnitType = 0
	UnitArcher   UnitType = 1
	UnitCavalry  UnitType = 2
)

// CounterMatrix[row][col] multiplier when attacker row hits defender col.
type CounterMatrix [3][3]fixed.Fixed

// DefaultCounters returns the embedded combat.json counter matrix.
func DefaultCounters() CounterMatrix {
	cfg, err := DefaultConfig()
	if err != nil {
		panic(err)
	}
	return cfg.Counters
}

// UnitStats holds deterministic combat numbers.
type UnitStats struct {
	ID      uint32
	Type    UnitType
	BaseHP  fixed.Fixed
	BaseATK fixed.Fixed
	BaseDEF fixed.Fixed
	HP      fixed.Fixed
	Move    int
	Range   int
}

// FinalATK computes attack after counter matrix.
func FinalATK(attacker UnitStats, defender UnitStats, counters CounterMatrix) fixed.Fixed {
	mul := counters[attacker.Type][defender.Type]
	return attacker.BaseATK.Mul(mul)
}

// Damage applies simple ATK - DEF clamped at zero (FP64).
func Damage(atk, def fixed.Fixed) fixed.Fixed {
	d := atk.Sub(def)
	if d.Raw() < 0 {
		return fixed.Zero
	}
	return d
}
