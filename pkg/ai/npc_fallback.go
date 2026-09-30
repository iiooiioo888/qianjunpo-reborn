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
	FallbackReasonHTTPStatus   FallbackReason = "http_non_ok_status"
	FallbackReasonBadResponse   FallbackReason = "bad_or_empty_response"
	FallbackReasonNoInferClient FallbackReason = "no_infer_client"
)

const (
	// SourceEdge marks a successful model response.
	SourceEdge = "edge"
	// SourceNPC marks persona template fallback text.
	SourceNPC = "npc"
)

// personaTemplates are deterministic NPC lines when edge/RAG inference is unavailable.
var personaTemplates = map[string]string{
	"guard":       "Hold the line!",
	"scout":       "Eyes on the pass—they're moving east.",
	"strategist":  "Wait for the signal, then strike together.",
	"merchant":    "Keep the convoy moving; we cannot linger.",
	"cao_cao":     "Strike where they least expect it.",
	"zhang_fei":   "Who dares cross this bridge?",
	"guan_yu":     "Cut their supply line before they rally.",
	"liu_bei":     "Stand with the people; hold the high ground.",
	"zhou_yu":     "Watch the wind—fire follows opportunity.",
	"sun_quan":    "Hold the river line; let none pass.",
}

// NPCFallback returns template dialogue for persona (normalized); unknown personas use the default line.
func NPCFallback(persona string) string {
	key := normalizePersona(persona)
	if line, ok := personaTemplates[key]; ok {
		return line
	}
	return "For the realm!"
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

func logNPCFallback(persona string, reason FallbackReason, detail string) {
	attrs := []any{
		slog.String("persona", persona),
		slog.String("reason", string(reason)),
		slog.String("source", SourceNPC),
	}
	if detail != "" {
		attrs = append(attrs, slog.String("detail", detail))
	}
	slog.Default().Info("ai infer npc fallback", attrs...)
}
