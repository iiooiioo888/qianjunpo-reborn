package balance

import (
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/combat"
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/fixed"
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/rng"
)

// Species mirrors combat unit type for win-rate reporting.
type Species = combat.UnitType

// MatchConfig holds stub combat parameters for batch simulation.
type MatchConfig struct {
	Seed     uint64
	AgentA   Agent
	AgentB   Agent
	Counters combat.CounterMatrix
}

// MatchResult is the outcome of one simulated match.
type MatchResult struct {
	Winner Species
	Rounds int
}

// Agent is the Phase 5 RL/MCTS pluggable policy interface (stub implementations in tests).
type Agent interface {
	Name() string
	// PickType chooses a species line-up for the stub duel.
	PickType(r *rng.RNG) Species
}

// StubAgent always picks a fixed species.
type StubAgent struct {
	Label string
	Type  Species
}

func (s StubAgent) Name() string { return s.Label }

func (s StubAgent) PickType(_ *rng.RNG) Species { return s.Type }

// SimulateMatch runs a deterministic rock-paper-scissors style duel using combat counters.
func SimulateMatch(cfg MatchConfig) MatchResult {
	r := rng.New(cfg.Seed)
	if cfg.Counters == ([3][3]fixed.Fixed{}) {
		cfg.Counters = combat.DefaultCounters()
	}
	aType := cfg.AgentA.PickType(r)
	bType := cfg.AgentB.PickType(r)
	winner := resolveWinner(aType, bType, cfg.Counters)
	return MatchResult{Winner: winner, Rounds: 1}
}

func resolveWinner(a, b Species, counters combat.CounterMatrix) Species {
	atk := combat.UnitStats{Type: a, BaseATK: fixed.FromInt(10), BaseDEF: fixed.FromInt(5)}
	def := combat.UnitStats{Type: b, BaseATK: fixed.FromInt(10), BaseDEF: fixed.FromInt(5)}
	dAB := combat.Damage(combat.FinalATK(atk, def, counters), def.BaseDEF)
	dBA := combat.Damage(combat.FinalATK(def, atk, counters), atk.BaseDEF)
	if dAB.Raw() == dBA.Raw() {
		if a <= b {
			return a
		}
		return b
	}
	if dAB.Raw() > dBA.Raw() {
		return a
	}
	return b
}
