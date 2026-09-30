package roma

import (
	"encoding/json"
	"errors"

	"github.com/iiooiioo888/qianjunpo-reborn/pkg/tactical"
)

// TacticalViewSnapshotJSON exports the live match as pkg/tactical.ViewSnapshot JSON.
func (s *Store) TacticalViewSnapshotJSON(id BattleID) ([]byte, uint64, uint64, error) {
	s.mu.RLock()
	b, ok := s.battles[id]
	s.mu.RUnlock()
	if !ok {
		return nil, 0, 0, errors.New("roma: battle not found")
	}
	if b.Match == nil {
		return nil, 0, 0, errors.New("roma: battle has no tactical match")
	}
	view := tactical.MatchToViewSnapshot(b.Match, s.timeFlowRateParts())
	data, err := json.Marshal(view)
	if err != nil {
		return nil, 0, 0, err
	}
	return data, b.Match.StateHash(), b.Match.Frame, nil
}
