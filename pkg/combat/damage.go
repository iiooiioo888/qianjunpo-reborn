package combat

import "github.com/iiooiioo888/qianjunpo-reborn/pkg/fixed"

// DamageRules selects how FinalATK and defender DEF combine into HP loss.
type DamageRules struct {
	// ArmorK enables ratio mitigation: dmg = max(0, atk * K / (K + def)).
	// Zero keeps legacy dmg = max(0, atk - def).
	ArmorK int64 `json:"armor_k"`
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
