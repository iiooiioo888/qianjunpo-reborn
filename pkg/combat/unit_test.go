package combat

import (
	"testing"

	"github.com/iiooiioo888/qianjunpo-reborn/pkg/fixed"
)

func TestCounters(t *testing.T) {
	c := DefaultCounters()
	inf := UnitStats{Type: UnitInfantry, BaseATK: fixed.FromInt(100)}
	arc := UnitStats{Type: UnitArcher, BaseDEF: fixed.FromInt(10)}
	atkStrong := FinalATK(inf, arc, c)
	infWeak := UnitStats{Type: UnitInfantry, BaseATK: fixed.FromInt(100)}
	cav := UnitStats{Type: UnitCavalry}
	atkWeak := FinalATK(infWeak, cav, c)
	if atkStrong.Raw() <= atkWeak.Raw() {
		t.Fatal("counter matrix should differentiate")
	}
}

