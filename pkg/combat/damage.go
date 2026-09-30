package combat

import "github.com/iiooiioo888/qianjunpo-reborn/pkg/fixed"

// DamageRules selects how FinalATK and defender DEF combine into HP loss.
type DamageRules struct {
	// ArmorK enables ratio mitigation: dmg = max(0, atk * K / (K + def)).
	// Zero keeps legacy dmg = max(0, atk - def).
	ArmorK int64 `json:"armor_k"`
	// HitRate: before crit, roll in [0, den); hit when roll < num. Zero num/den disables miss roll (always hit).
	HitRate Ratio `json:"hit_rate"`
	// CritRate: on hit, roll in [0, den); crit when roll < num. Zero num/den disables crit.
	CritRate Ratio `json:"crit_rate"`
	// CritMul multiplies post-mitigation damage on crit (e.g. 150/100). Ignored when crit disabled.
	CritMul Ratio `json:"crit_mul"`
}

// Damage applies simple ATK - DEF clamped at zero (FP64).
func Damage(atk, def fixed.Fixed) fixed.Fixed {
	d := atk.Sub(def)
	if d.Raw() < 0 {
		return fixed.Zero
	}
	return d
}

// DamageArmor applies deterministic ratio DR: atk * K / (K + def), clamped at zero.
func DamageArmor(atk, def fixed.Fixed, k fixed.Fixed) fixed.Fixed {
	if atk.Raw() <= 0 || k.Raw() <= 0 {
		return fixed.Zero
	}
	denom := k.Add(def)
	if denom.Raw() <= 0 {
		return fixed.Zero
	}
	return atk.Mul(k).Div(denom)
}

// ResolveDamage picks the configured model (default subtract).
func (c Config) ResolveDamage(atk, def fixed.Fixed) fixed.Fixed {
	if c.Damage.ArmorK <= 0 {
		return Damage(atk, def)
	}
	return DamageArmor(atk, def, fixed.FromInt(c.Damage.ArmorK))
}

// HitEnabled reports whether hit_rate is configured for this rules snapshot.
func (c Config) HitEnabled() bool {
	return c.Damage.HitRate.Num > 0 && c.Damage.HitRate.Den > 0
}

// CritEnabled reports whether crit_rate is configured for this rules snapshot.
func (c Config) CritEnabled() bool {
	return c.Damage.CritRate.Num > 0 && c.Damage.CritRate.Den > 0
}

// RollHit is deterministic: roll mod den < num (same rule shape as crit).
func RollHit(roll uint64, rate Ratio) bool {
	return rollPass(roll, rate)
}

// RollCrit is deterministic: roll mod den < num.
func RollCrit(roll uint64, rate Ratio) bool {
	return rollPass(roll, rate)
}

func rollPass(roll uint64, rate Ratio) bool {
	if rate.Num <= 0 || rate.Den <= 0 {
		return false
	}
	return roll%uint64(rate.Den) < uint64(rate.Num)
}

// ApplyHitToDamage returns zero on miss when hit_rate is enabled; otherwise returns dmg unchanged.
func (c Config) ApplyHitToDamage(dmg fixed.Fixed, roll uint64) fixed.Fixed {
	if !c.HitEnabled() || RollHit(roll, c.Damage.HitRate) {
		return dmg
	}
	return fixed.Zero
}

// ApplyCritMultiplier scales damage by crit_mul; invalid mul returns dmg unchanged.
func ApplyCritMultiplier(dmg fixed.Fixed, mul Ratio) fixed.Fixed {
	if dmg.Raw() <= 0 || mul.Num <= 0 || mul.Den <= 0 {
		return dmg
	}
	fx, err := mul.ToFixed()
	if err != nil {
		return dmg
	}
	return dmg.Mul(fx)
}

// ApplyCritToDamage optionally multiplies dmg when roll passes crit_rate.
func (c Config) ApplyCritToDamage(dmg fixed.Fixed, roll uint64) fixed.Fixed {
	if !c.CritEnabled() || !RollCrit(roll, c.Damage.CritRate) {
		return dmg
	}
	return ApplyCritMultiplier(dmg, c.Damage.CritMul)
}
