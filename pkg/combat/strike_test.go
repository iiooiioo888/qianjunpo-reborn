package combat

import (
	"testing"

	"github.com/iiooiioo888/qianjunpo-reborn/pkg/fixed"
)

func TestResolveStrikeDamageMatchesSubtractPipeline(t *testing.T) {
	cfg, err := DefaultConfig()
	if err != nil {
		t.Fatal(err)
	}
	att := UnitStats{Type: UnitInfantry, BaseATK: fixed.FromInt(20)}
	def := UnitStats{Type: UnitCavalry, BaseDEF: fixed.FromInt(5)}
	got := ResolveStrikeDamage(cfg, cfg.Counters, att, def, 0, 0)
	atk := FinalATK(att, def, cfg.Counters)
	want := cfg.ResolveDamage(atk, def.BaseDEF)
	if got.Raw() != want.Raw() {
		t.Fatalf("got %d want %d", got.Raw(), want.Raw())
	}
}
