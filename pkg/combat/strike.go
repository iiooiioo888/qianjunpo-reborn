package combat

import "github.com/iiooiioo888/qianjunpo-reborn/pkg/fixed"

// ResolveStrikeDamage applies the configured mitigation, hit, and crit pipeline
// for one attacker→defender pair. Rolls are consumed only when hit_rate / crit_rate
// are enabled on cfg (same semantics as tactical single-target attacks).
func ResolveStrikeDamage(cfg Config, counters CounterMatrix, attacker, defender UnitStats, hitRoll, critRoll uint64) fixed.Fixed {
	atk := FinalATK(attacker, defender, counters)
	dmg := cfg.ResolveDamage(atk, defender.BaseDEF)
	if cfg.HitEnabled() {
		dmg = cfg.ApplyHitToDamage(dmg, hitRoll)
	}
	if cfg.CritEnabled() && dmg.Raw() > 0 {
		dmg = cfg.ApplyCritToDamage(dmg, critRoll)
	}
	return dmg
}
