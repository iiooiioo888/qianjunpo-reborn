package combat

import "github.com/iiooiioo888/qianjunpo-reborn/pkg/fixed"

// SkillID identifies a tactical skill or card (Phase 2+ hook).
type SkillID uint16

// CardPoolID identifies a deck pool for matchmaking / AI (stub).
type CardPoolID uint32

// SkillModifier is a placeholder hook to adjust FinalATK before damage (not wired in sim yet).
type SkillModifier func(attacker, defender UnitStats) fixed.Fixed

// Stub skill ids for the Phase 2 vertical slice (behavior wired in pkg/tactical).
const (
	SkillStubStrike SkillID = 1 // single-target strike at Command.To (KindAttack rules)
	SkillStubSplash SkillID = 2 // Chebyshev splash at Command.To (KindAoE rules)
)

const DefaultCardPoolID CardPoolID = 1

// SkillBehavior tells tactical how to validate/resolve a skill (stub catalog only).
type SkillBehavior uint8

const (
	SkillBehaviorStrike SkillBehavior = 1
	SkillBehaviorSplash SkillBehavior = 2
)

// SkillDef is a minimal catalog entry (no counter-table changes).
type SkillDef struct {
	ID       SkillID
	Behavior SkillBehavior
}

// CardPool lists skills legal in a match deck (submit-time gate).
type CardPool struct {
	ID     CardPoolID
	Skills []SkillID
}

var defaultSkillCatalog = map[SkillID]SkillDef{
	SkillStubStrike: {ID: SkillStubStrike, Behavior: SkillBehaviorStrike},
	SkillStubSplash: {ID: SkillStubSplash, Behavior: SkillBehaviorSplash},
}

var defaultCardPool = CardPool{
	ID:     DefaultCardPoolID,
	Skills: []SkillID{SkillStubStrike, SkillStubSplash},
}

// DefaultCardPool returns the built-in duel pool used by tactical.NewMatch.
func DefaultCardPool() CardPool {
	return defaultCardPool
}

// LookupSkill returns catalog metadata for a skill id.
func LookupSkill(id SkillID) (SkillDef, bool) {
	def, ok := defaultSkillCatalog[id]
	return def, ok
}

// Contains reports whether id is in the pool (linear scan; stub deck sizes only).
func (p CardPool) Contains(id SkillID) bool {
	for _, sid := range p.Skills {
		if sid == id {
			return true
		}
	}
	return false
}
