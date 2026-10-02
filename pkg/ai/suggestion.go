package ai

import (
	"strings"

	"github.com/iiooiioo888/qianjunpo-reborn/pkg/tactical"
)

// InferPath labels which stagger leg produced the suggestion (edge-only vs RAG-augmented).
const (
	InferPathEdge = "edge"
	InferPathRAG  = "rag"
)

// SuggestionKind mirrors tactical command kinds for HUD/curl (move or skill, never both).
const (
	SuggestionKindMove  = "move"
	SuggestionKindSkill = "skill"
)

// InferSuggestion is one consumable tactical hint (unit + target grid); not submitted to Roma.
type InferSuggestion struct {
	Kind     string `json:"kind"` // move | skill
	UnitID   uint32 `json:"unit_id"`
	ToX      int    `json:"to_x"`
	ToY      int    `json:"to_y"`
	SkillID  uint32 `json:"skill_id,omitempty"` // KindSkill only
	InferPath string `json:"infer_path"`        // edge | rag
}

// DeriveSuggestion maps model text (+ optional RAG hits) into a single move or skill hint.
// Deterministic for mock/Ollama stubs; safe on NPC template text (defaults to a demo move).
func DeriveSuggestion(path string, text string, ragHitIDs []string) InferSuggestion {
	if path != InferPathRAG {
		path = InferPathEdge
	}
	lower := strings.ToLower(text)
	if path == InferPathRAG {
		if ragHas(ragHitIDs, "tactic-spear") || strings.Contains(lower, "spear") {
			return skillSuggestion(path, 9, 9)
		}
		if ragHas(ragHitIDs, "tactic-flank") || strings.Contains(lower, "flank") {
			return moveSuggestion(path, 8, 8)
		}
	}
	if strings.Contains(lower, "flank") {
		return moveSuggestion(path, 8, 8)
	}
	if strings.Contains(lower, "skill") || strings.Contains(lower, "strike") {
		return skillSuggestion(path, 9, 9)
	}
	return moveSuggestion(path, 5, 8)
}

func moveSuggestion(path string, x, y int) InferSuggestion {
	return InferSuggestion{
		Kind:      SuggestionKindMove,
		UnitID:    tactical.UnitIDPlayer0,
		ToX:       x,
		ToY:       y,
		InferPath: path,
	}
}

func skillSuggestion(path string, x, y int) InferSuggestion {
	return InferSuggestion{
		Kind:      SuggestionKindSkill,
		UnitID:    tactical.UnitIDPlayer0,
		ToX:       x,
		ToY:       y,
		SkillID:   1, // combat.SkillStubStrike; HUD maps to POST kind=5
		InferPath: path,
	}
}

func ragHas(ids []string, want string) bool {
	for _, id := range ids {
		if id == want {
			return true
		}
	}
	return false
}
