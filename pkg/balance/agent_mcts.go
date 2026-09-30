package balance

import "github.com/iiooiioo888/qianjunpo-reborn/pkg/rng"

// MCTSAgent is a placeholder Monte-Carlo tree search policy (Phase 5 skeleton).
type MCTSAgent struct {
	NameLabel string
	Simulations int
}

func (m MCTSAgent) Name() string {
	if m.NameLabel != "" {
		return m.NameLabel
	}
	return "mcts-stub"
}

// PickType uses deterministic RNG rollouts (no tree in CI).
func (m MCTSAgent) PickType(r *rng.RNG) Species {
	_ = m.Simulations
	roll := r.NextUint64() % 3
	return Species(roll)
}
