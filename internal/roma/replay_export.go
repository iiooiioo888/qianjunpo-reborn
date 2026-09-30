package roma

import (
	"errors"

	"github.com/iiooiioo888/qianjunpo-reborn/pkg/replay"
)

// ExportRecording returns the pkg/replay gzip-compatible recording for a battle.
// It seals the in-memory match recording when the battle has finished stepping.
func (s *Store) ExportRecording(id BattleID) (replay.Recording, error) {
	s.mu.RLock()
	b, ok := s.battles[id]
	s.mu.RUnlock()
	if !ok {
		return replay.Recording{}, errors.New("roma: battle not found")
	}
	if b.Match == nil {
		return replay.Recording{}, errors.New("roma: battle has no tactical match")
	}
	if !b.Match.Finished {
		return replay.Recording{}, errors.New("roma: battle not finished")
	}
	b.Match.FinishRecording()
	rec := b.Match.Recording()
	if err := replay.VerifyConsistency(rec); err != nil {
		return replay.Recording{}, err
	}
	return rec, nil
}
