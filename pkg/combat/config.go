package combat

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/iiooiioo888/qianjunpo-reborn/pkg/fixed"
)

//go:embed default_combat.json
var embeddedCombatJSON []byte

// Ratio is an exact rational multiplier (num/den), loaded without float on the sim path.
type Ratio struct {
	Num int64 `json:"num"`
	Den int64 `json:"den"`
}

func (r Ratio) ToFixed() (fixed.Fixed, error) {
	if r.Den == 0 {
		return fixed.Zero, fmt.Errorf("combat: ratio denominator zero")
	}
	return fixed.FromInt(r.Num).Div(fixed.FromInt(r.Den)), nil
}

// UnitCatalogEntry is static data for one unit type.
type UnitCatalogEntry struct {
	BaseHP    int64 `json:"base_hp"`
	BaseATK   int64 `json:"base_atk"`
	BaseDEF   int64 `json:"base_def"`
	Move      int   `json:"move"`
	Range     int   `json:"range"`
	CostFood  int64 `json:"cost_food"`
	CostGold  int64 `json:"cost_gold"`
}

// Config is the combat rules snapshot used for one match.
type Config struct {
	Version  int                           `json:"version"`
	Counters CounterMatrix                 `json:"-"`
	Units    map[UnitType]UnitCatalogEntry `json:"-"`
	rawUnits map[string]UnitCatalogEntry   `json:"-"`
	rawMatrix [3][3]Ratio                  `json:"-"`
}

type configJSON struct {
	Version  int                         `json:"version"`
	Counters struct {
		Matrix [3][3]Ratio `json:"matrix"`
	} `json:"counters"`
	Units map[string]UnitCatalogEntry `json:"units"`
}

// DefaultConfig returns the embedded combat.json compiled into the binary.
func DefaultConfig() (Config, error) {
	return LoadBytes(embeddedCombatJSON)
}

// LoadFile reads combat rules from a JSON file (init / match lobby only).
func LoadFile(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}
	return LoadBytes(data)
}

// LoadBytes parses combat JSON and materializes fixed-point counters.
func LoadBytes(data []byte) (Config, error) {
	var raw configJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return Config{}, fmt.Errorf("combat: parse config: %w", err)
	}
	cfg := Config{
		Version:  raw.Version,
		rawUnits: raw.Units,
	}
	if cfg.rawUnits == nil {
		cfg.rawUnits = make(map[string]UnitCatalogEntry)
	}
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			cfg.rawMatrix[i][j] = raw.Counters.Matrix[i][j]
			fx, err := raw.Counters.Matrix[i][j].ToFixed()
			if err != nil {
				return Config{}, fmt.Errorf("combat: counter[%d][%d]: %w", i, j, err)
			}
			cfg.Counters[i][j] = fx
		}
	}
	cfg.Units = make(map[UnitType]UnitCatalogEntry, len(raw.Units))
	for name, entry := range raw.Units {
		typ, err := parseUnitType(name)
		if err != nil {
			return Config{}, err
		}
		if entry.Range < 1 {
			return Config{}, fmt.Errorf("combat: %s range must be >= 1", name)
		}
		if entry.Move < 0 {
			return Config{}, fmt.Errorf("combat: %s move must be >= 0", name)
		}
		cfg.Units[typ] = entry
	}
	for _, typ := range []UnitType{UnitInfantry, UnitArcher, UnitCavalry} {
		if _, ok := cfg.Units[typ]; !ok {
			return Config{}, fmt.Errorf("combat: missing unit catalog for type %d", typ)
		}
	}
	return cfg, nil
}

func parseUnitType(name string) (UnitType, error) {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "infantry":
		return UnitInfantry, nil
	case "archer":
		return UnitArcher, nil
	case "cavalry":
		return UnitCavalry, nil
	default:
		return 0, fmt.Errorf("combat: unknown unit type %q", name)
	}
}

// CatalogStats builds runtime UnitStats (current HP = base HP) for a type.
func (c Config) CatalogStats(id uint32, typ UnitType) (UnitStats, error) {
	entry, ok := c.Units[typ]
	if !ok {
		return UnitStats{}, fmt.Errorf("combat: no catalog for type %d", typ)
	}
	return UnitStats{
		ID:      id,
		Type:    typ,
		BaseHP:  fixed.FromInt(entry.BaseHP),
		BaseATK: fixed.FromInt(entry.BaseATK),
		BaseDEF: fixed.FromInt(entry.BaseDEF),
		HP:      fixed.FromInt(entry.BaseHP),
		Move:    entry.Move,
		Range:   entry.Range,
	}, nil
}

// MarshalJSON round-trips the logical config (for tests).
func (c Config) MarshalJSON() ([]byte, error) {
	raw := configJSON{Version: c.Version}
	if len(c.rawUnits) > 0 {
		raw.Units = c.rawUnits
	} else {
		raw.Units = make(map[string]UnitCatalogEntry)
		for typ, entry := range c.Units {
			raw.Units[unitTypeName(typ)] = entry
		}
	}
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			raw.Counters.Matrix[i][j] = c.rawMatrix[i][j]
		}
	}
	return json.Marshal(raw)
}

func unitTypeName(t UnitType) string {
	switch t {
	case UnitInfantry:
		return "infantry"
	case UnitArcher:
		return "archer"
	case UnitCavalry:
		return "cavalry"
	default:
		return fmt.Sprintf("unknown_%d", t)
	}
}

