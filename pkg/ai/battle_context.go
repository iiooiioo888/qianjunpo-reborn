package ai

import "strings"

// BattleContext is a compact live-battle summary for RAG retrieval and prompt augmentation.
// It is display/strategic only and does not enter FP64 combat paths.
type BattleContext struct {
	UnitsSummary   string `json:"units_summary,omitempty"`
	TerrainSummary string `json:"terrain_summary,omitempty"`
}

// RetrievalQuery combines battle summary with the player's order for vector Top-K.
func (b BattleContext) RetrievalQuery(order string) string {
	var parts []string
	if s := strings.TrimSpace(b.UnitsSummary); s != "" {
		parts = append(parts, "units: "+s)
	}
	if s := strings.TrimSpace(b.TerrainSummary); s != "" {
		parts = append(parts, "terrain: "+s)
	}
	if s := strings.TrimSpace(order); s != "" {
		parts = append(parts, s)
	}
	return strings.Join(parts, " ")
}

// PromptBlock formats the live situation for the LLM prompt (after RAG knowledge).
func (b BattleContext) PromptBlock() string {
	var lines []string
	if s := strings.TrimSpace(b.UnitsSummary); s != "" {
		lines = append(lines, "Units: "+s)
	}
	if s := strings.TrimSpace(b.TerrainSummary); s != "" {
		lines = append(lines, "Terrain: "+s)
	}
	if len(lines) == 0 {
		return ""
	}
	return "Battlefield:\n" + strings.Join(lines, "\n")
}
