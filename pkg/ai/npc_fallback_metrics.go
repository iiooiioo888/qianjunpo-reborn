package ai

import "sync"

var npcFallbackCounter npcFallbackCounterState

type npcFallbackCounterState struct {
	mu       sync.Mutex
	byReason map[string]uint64
}

func recordNPCFallback(reason FallbackReason) {
	if reason == FallbackReasonNone {
		return
	}
	key := string(reason)
	npcFallbackCounter.mu.Lock()
	if npcFallbackCounter.byReason == nil {
		npcFallbackCounter.byReason = make(map[string]uint64)
	}
	npcFallbackCounter.byReason[key]++
	npcFallbackCounter.mu.Unlock()
}

// NPCFallbackCounts returns in-process NPC fallback totals keyed by fallback_reason.
func NPCFallbackCounts() map[string]uint64 {
	npcFallbackCounter.mu.Lock()
	defer npcFallbackCounter.mu.Unlock()
	if len(npcFallbackCounter.byReason) == 0 {
		return map[string]uint64{}
	}
	out := make(map[string]uint64, len(npcFallbackCounter.byReason))
	for k, v := range npcFallbackCounter.byReason {
		out[k] = v
	}
	return out
}

// ResetNPCFallbackCounts clears counters (tests only).
func ResetNPCFallbackCounts() {
	npcFallbackCounter.mu.Lock()
	npcFallbackCounter.byReason = make(map[string]uint64)
	npcFallbackCounter.mu.Unlock()
}
