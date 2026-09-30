package ai

import (
	"log/slog"
	"strings"
)

// FallbackReason explains why NPC template text was returned instead of edge inference.
type FallbackReason string

const (
	FallbackReasonNone          FallbackReason = ""
	FallbackReasonRequestBuild  FallbackReason = "request_build_failed"
	FallbackReasonHTTPTransport FallbackReason = "http_transport_error"
	FallbackReasonTimeout       FallbackReason = "timeout"
	FallbackReasonHTTPStatus    FallbackReason = "http_non_ok_status"
	FallbackReasonBadResponse   FallbackReason = "bad_or_empty_response"
	FallbackReasonNoInferClient FallbackReason = "no_infer_client"
)

const (
	// SourceEdge marks a successful model response.
	SourceEdge = "edge"
	// SourceNPC marks persona template fallback text.
	SourceNPC = "npc"
)

// RegistryPersonaKeys are canonical persona ids aligned with client v04 character cards
// (CharacterCardSpriteRegistry) and unit sprites (UnitSpriteRegistry).
var RegistryPersonaKeys = []string{
	"char_caocao",
	"char_zhangfei",
	"char_wu_placeholder",
	"infantry",
	"cavalry",
}

// personaTemplates are deterministic NPC lines when edge/RAG inference is unavailable.
// Primary keys match client registry ids; generic tactical roles remain for non-card NPCs.
var personaTemplates = map[string]string{
	"char_caocao":         "Strike where they least expect it.",
	"char_zhangfei":       "Who dares cross this bridge?",
	"char_wu_placeholder": "Watch the wind—fire follows opportunity.",
	"infantry":            "Form the line—hold every step.",
	"cavalry":             "Ride the flank before they set their spears.",
	"guard":               "Hold the line!",
	"scout":               "Eyes on the pass—they're moving east.",
	"strategist":          "Wait for the signal, then strike together.",
	"merchant":            "Keep the convoy moving; we cannot linger.",
	// Legacy general ids (no v04 card yet); kept for older callers.
	"guan_yu":  "Cut their supply line before they rally.",
	"liu_bei":  "Stand with the people; hold the high ground.",
	"zhou_yu":  "Watch the wind—fire follows opportunity.",
	"sun_quan": "Hold the river line; let none pass.",
}

// personaAliases maps normalized legacy or shorthand ids to registry-aligned keys.
var personaAliases = map[string]string{
	"cao_cao":   "char_caocao",
	"zhang_fei": "char_zhangfei",
	"zhangfei":  "char_zhangfei",
	"caocao":    "char_caocao",
	"wu_placeholder": "char_wu_placeholder",
	"char_wu":   "char_wu_placeholder",
}

// NPCFallback returns template dialogue for persona (normalized); unknown personas use the default line.
func NPCFallback(persona string) string {
	key := resolvePersonaKey(persona)
	if line, ok := personaTemplates[key]; ok {
		return line
	}
	return "For the realm!"
}

func resolvePersonaKey(persona string) string {
	key := normalizePersona(persona)
	if canon, ok := personaAliases[key]; ok {
		return canon
	}
	return key
}

func normalizePersona(persona string) string {
	p := strings.TrimSpace(strings.ToLower(persona))
	p = strings.ReplaceAll(p, "-", "_")
	p = strings.ReplaceAll(p, " ", "_")
	return p
}

func npcResult(persona string, reason FallbackReason, detail string) InferResult {
	return InferResult{
		Text:           NPCFallback(persona),
		Source:         SourceNPC,
		FallbackReason: reason,
		FallbackDetail: detail,
	}
}

func edgeResult(text string) InferResult {
	return InferResult{
		Text:   text,
		Source: SourceEdge,
	}
}

// logNPCFallback records observability fields aligned with InferResult JSON (source, fallback_reason, fallback_detail).
func logNPCFallback(persona string, reason FallbackReason, detail string) {
	attrs := []any{
		slog.String("source", SourceNPC),
		slog.String("fallback_reason", string(reason)),
	}
	if detail != "" {
		attrs = append(attrs, slog.String("fallback_detail", detail))
	}
	if persona != "" {
		attrs = append(attrs, slog.String("persona", persona))
	}
	slog.Default().Info("ai infer npc fallback", attrs...)
}

func npcFallbackResult(persona string, reason FallbackReason, detail string, log bool) InferResult {
	recordNPCFallback(reason)
	if log {
		logNPCFallback(persona, reason, detail)
	}
	return npcResult(persona, reason, detail)
}
