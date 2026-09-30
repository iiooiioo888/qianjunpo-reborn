package combat

import (
	"testing"

	"github.com/iiooiioo888/qianjunpo-reborn/pkg/fixed"
)

func TestDamageNonNegative(t *testing.T) {
	d := Damage(fixed.FromInt(5), fixed.FromInt(10))
	if d.Raw() != 0 {
		t.Fatal("damage clamped")
	}
}

func TestDamageArmorRatio(t *testing.T) {
	atk := fixed.FromInt(40)
	def := fixed.FromInt(10)
	k := fixed.FromInt(100)
	got := DamageArmor(atk, def, k)
	// 40 * 100 / (100 + 10) = 4000/110 ≈ 36.363...
	want := fixed.FromInt(36).Add(fixed.FromInt(4).Div(fixed.FromInt(11)))
	if got.Raw() != want.Raw() {
		t.Fatalf("DamageArmor golden: got raw %d want %d", got.Raw(), want.Raw())
	}
}

func TestResolveDamageDefaultSubtract(t *testing.T) {
	cfg, err := DefaultConfig()
	if err != nil {
		t.Fatal(err)
	}
	atk := fixed.FromInt(25)
	def := fixed.FromInt(10)
	if cfg.Damage.ArmorK != 0 {
		t.Fatalf("expected default subtract mode, armor_k=%d", cfg.Damage.ArmorK)
	}
	got := cfg.ResolveDamage(atk, def)
	want := Damage(atk, def)
	if got.Raw() != want.Raw() {
		t.Fatalf("ResolveDamage default: got %d want %d", got.Raw(), want.Raw())
	}
}

func TestResolveDamageArmorFromConfig(t *testing.T) {
	data := []byte(`{
  "version": 1,
  "damage": { "armor_k": 100 },
  "counters": {
    "matrix": [
      [{"num": 100, "den": 100}, {"num": 100, "den": 100}, {"num": 100, "den": 100}],
      [{"num": 100, "den": 100}, {"num": 100, "den": 100}, {"num": 100, "den": 100}],
      [{"num": 100, "den": 100}, {"num": 100, "den": 100}, {"num": 100, "den": 100}]
    ]
  },
  "units": {
    "infantry": {"base_hp": 1, "base_atk": 1, "base_def": 1, "move": 1, "range": 1, "cost_food": 0, "cost_gold": 0},
    "archer": {"base_hp": 1, "base_atk": 1, "base_def": 1, "move": 1, "range": 1, "cost_food": 0, "cost_gold": 0},
    "cavalry": {"base_hp": 1, "base_atk": 1, "base_def": 1, "move": 1, "range": 1, "cost_food": 0, "cost_gold": 0}
  }
}`)
	cfg, err := LoadBytes(data)
	if err != nil {
		t.Fatal(err)
	}
	atk := fixed.FromInt(40)
	def := fixed.FromInt(10)
	got := cfg.ResolveDamage(atk, def)
	want := DamageArmor(atk, def, fixed.FromInt(100))
	if got.Raw() != want.Raw() {
		t.Fatalf("ResolveDamage armor: got %d want %d", got.Raw(), want.Raw())
	}
}

func TestDefaultConfigCritDisabled(t *testing.T) {
	cfg, err := DefaultConfig()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.CritEnabled() {
		t.Fatal("embedded default must not enable crit")
	}
	base := fixed.FromInt(20)
	if got := cfg.ApplyCritToDamage(base, 0); got.Raw() != base.Raw() {
		t.Fatalf("crit off: got %d want %d", got.Raw(), base.Raw())
	}
}

func TestRollCritDeterministic(t *testing.T) {
	rate := Ratio{Num: 1, Den: 4}
	if !RollCrit(0, rate) {
		t.Fatal("roll 0 should crit")
	}
	if RollCrit(1, rate) {
		t.Fatal("roll 1 should not crit")
	}
	if RollCrit(3, rate) {
		t.Fatal("roll 3 should not crit")
	}
}

func TestApplyCritToDamageFromConfig(t *testing.T) {
	data := []byte(`{
  "version": 1,
  "damage": {
    "crit_rate": {"num": 1, "den": 1},
    "crit_mul": {"num": 150, "den": 100}
  },
  "counters": {
    "matrix": [
      [{"num": 100, "den": 100}, {"num": 100, "den": 100}, {"num": 100, "den": 100}],
      [{"num": 100, "den": 100}, {"num": 100, "den": 100}, {"num": 100, "den": 100}],
      [{"num": 100, "den": 100}, {"num": 100, "den": 100}, {"num": 100, "den": 100}]
    ]
  },
  "units": {
    "infantry": {"base_hp": 1, "base_atk": 1, "base_def": 1, "move": 1, "range": 1, "cost_food": 0, "cost_gold": 0},
    "archer": {"base_hp": 1, "base_atk": 1, "base_def": 1, "move": 1, "range": 1, "cost_food": 0, "cost_gold": 0},
    "cavalry": {"base_hp": 1, "base_atk": 1, "base_def": 1, "move": 1, "range": 1, "cost_food": 0, "cost_gold": 0}
  }
}`)
	cfg, err := LoadBytes(data)
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.CritEnabled() {
		t.Fatal("expected crit enabled")
	}
	base := fixed.FromInt(20)
	got := cfg.ApplyCritToDamage(base, 0)
	want := fixed.FromInt(30) // 20 * 1.5
	if got.Raw() != want.Raw() {
		t.Fatalf("crit dmg: got %d want %d", got.Raw(), want.Raw())
	}
}
