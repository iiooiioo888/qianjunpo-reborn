package combat

import "github.com/iiooiioo888/qianjunpo-reborn/pkg/fixed"

// SkillID identifies a tactical skill or card (Phase 2+ hook only).
type SkillID uint16

// CardPoolID identifies a deck pool for matchmaking / AI (stub).
type CardPoolID uint32

// SkillModifier is a placeholder hook to adjust FinalATK before damage (not wired in sim yet).
type SkillModifier func(attacker, defender UnitStats) fixed.Fixed
