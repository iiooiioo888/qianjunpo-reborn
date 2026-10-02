package ai

import (
	"sync"
)

// SuggestionStore holds the latest infer suggestion per battle_id for curl/UI polling.
type SuggestionStore struct {
	mu   sync.RWMutex
	byID map[string]InferSuggestion
}

var defaultSuggestionStore = &SuggestionStore{byID: make(map[string]InferSuggestion)}

// GlobalSuggestionStore is the process-wide write-back target (edge-infer / gateways).
func GlobalSuggestionStore() *SuggestionStore {
	return defaultSuggestionStore
}

// WriteSuggestion stores one suggestion for battleID (overwrites prior).
func (s *SuggestionStore) WriteSuggestion(battleID string, sug InferSuggestion) {
	if battleID == "" {
		return
	}
	s.mu.Lock()
	s.byID[battleID] = sug
	s.mu.Unlock()
}

// GetSuggestion returns the last write-back for battleID.
func (s *SuggestionStore) GetSuggestion(battleID string) (InferSuggestion, bool) {
	s.mu.RLock()
	sug, ok := s.byID[battleID]
	s.mu.RUnlock()
	return sug, ok
}

// ResetSuggestions clears all entries (tests).
func (s *SuggestionStore) ResetSuggestions() {
	s.mu.Lock()
	s.byID = make(map[string]InferSuggestion)
	s.mu.Unlock()
}
