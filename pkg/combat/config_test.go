package combat

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/iiooiioo888/qianjunpo-reborn/pkg/fixed"
)

func TestLoadFileMatchesEmbeddedDefault(t *testing.T) {
	def, err := DefaultConfig()
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join("..", "..", "config", "combat", "combat.json")
	fileCfg, err := LoadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if def.Counters != fileCfg.Counters {
		t.Fatal("counter matrix mismatch between embed and config/combat/combat.json")
	}
	for _, typ := range []UnitType{UnitInfantry, UnitArcher, UnitCavalry} {
		if def.Units[typ] != fileCfg.Units[typ] {
			t.Fatalf("catalog mismatch type %d", typ)
		}
	}
}

func TestConfigJSONRoundTrip(t *testing.T) {
	cfg, err := DefaultConfig()
	if err != nil {
		t.Fatal(err)
	}
	data, err := cfg.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}
	again, err := LoadBytes(data)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Counters != again.Counters {
		t.Fatal("counter round-trip mismatch")
	}
}

func TestFinalATKGoldenInfantryVsArcher(t *testing.T) {
	cfg, err := DefaultConfig()
	if err != nil {
		t.Fatal(err)
	}
	inf, err := cfg.CatalogStats(1, UnitInfantry)
	if err != nil {
		t.Fatal(err)
	}
	arc, err := cfg.CatalogStats(2, UnitArcher)
	if err != nil {
		t.Fatal(err)
	}
	atk := FinalATK(inf, arc, cfg.Counters)
	// 30 * 1.25 = 37.5 (FP64 32.32)
	want := fixed.FromInt(37).Add(fixed.FromInt(1).Div(fixed.FromInt(2)))
	if atk.Raw() != want.Raw() {
		t.Fatalf("FinalATK golden: got raw %d want %d", atk.Raw(), want.Raw())
	}
}

func TestLoadFileFromRepoConfig(t *testing.T) {
	if _, err := os.Stat(filepath.Join("config", "combat", "combat.json")); err != nil {
		t.Skip("repo config path not found from package dir")
	}
	_, err := LoadFile(filepath.Join("config", "combat", "combat.json"))
	if err != nil {
		t.Fatal(err)
	}
}
