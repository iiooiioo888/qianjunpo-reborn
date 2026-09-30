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

// DefaultCounters returns brief-style rock-paper-scissors style modifiers (FP64).
func DefaultCounters() CounterMatrix {
	// Base 1.0, strong 1.25, weak 0.8 — all via fixed.FromFloat at init only.
	one := fixed.FromInt(1)
	strong := fixed.FromFloat(1.25)
	weak := fixed.FromFloat(0.8)
	m := CounterMatrix{}
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			m[i][j] = one
		}
	}
	// Infantry > Archer, Archer > Cavalry, Cavalry > Infantry
	m[UnitInfantry][UnitArcher] = strong
	m[UnitInfantry][UnitCavalry] = weak
	m[UnitArcher][UnitCavalry] = strong
	m[UnitArcher][UnitInfantry] = weak
	m[UnitCavalry][UnitInfantry] = strong
	m[UnitCavalry][UnitArcher] = weak
	return m
}

// UnitStats holds deterministic combat numbers.
type UnitStats struct {
	ID       uint32
	Type     UnitType
	BaseATK  fixed.Fixed
	BaseDEF  fixed.Fixed
	HP       fixed.Fixed
	Move     int
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
