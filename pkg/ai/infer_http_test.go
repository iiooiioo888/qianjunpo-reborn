package ai

import (
	"encoding/json"
	"testing"
)

func TestEdgeInferHTTPResponseJSON(t *testing.T) {
	raw, err := json.Marshal(EdgeInferHTTPResponse("advance: gate", "qwen2.5-3b-mock", 3))
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatal(err)
	}
	if m["source"] != SourceEdge {
		t.Fatalf("source=%v", m["source"])
	}
	if _, ok := m["fallback_reason"]; ok {
		t.Fatalf("fallback_reason should be omitted: %v", m)
	}
	if _, ok := m["fallback_detail"]; ok {
		t.Fatalf("fallback_detail should be omitted: %v", m)
	}
}

func TestInferResultFromHTTPBackwardCompat(t *testing.T) {
	out := InferResultFromHTTP(InferHTTPResponse{Text: "ok", Model: "m", LatencyMs: 1})
	if out.Source != SourceEdge || out.FallbackReason != FallbackReasonNone || out.Text != "ok" {
		t.Fatalf("%+v", out)
	}
	if out.RAGK != 0 || len(out.RAGHitIDs) != 0 {
		t.Fatalf("rag fields should be empty: %+v", out)
	}
}

func TestInferResultFromHTTPRAGFields(t *testing.T) {
	out := InferResultFromHTTP(InferHTTPResponse{
		Text:      "line",
		Model:     "m",
		LatencyMs: 2,
		Source:    SourceEdge,
		RAGK:      5,
		RAGHitIDs: []string{"tactic-flank", "wei-cao"},
	})
	if out.RAGK != 5 || len(out.RAGHitIDs) != 2 || out.RAGHitIDs[0] != "tactic-flank" {
		t.Fatalf("%+v", out)
	}
}

func TestRAGInferHTTPResponseJSON(t *testing.T) {
	raw, err := json.Marshal(RAGInferHTTPResponse("line", "m", 2, 5, []string{"tactic-flank", "wei-xiahou"}))
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatal(err)
	}
	if m["rag_k"] != float64(5) {
		t.Fatalf("rag_k=%v", m["rag_k"])
	}
	ids, ok := m["rag_hit_ids"].([]interface{})
	if !ok || len(ids) != 2 {
		t.Fatalf("rag_hit_ids=%v", m["rag_hit_ids"])
	}
}

func TestInferResultFromHTTPNPCOnWire(t *testing.T) {
	out := InferResultFromHTTP(InferHTTPResponse{
		Text:           "Hold the line!",
		Source:         SourceNPC,
		FallbackReason: string(FallbackReasonTimeout),
		FallbackDetail: "edge proxy timeout",
	})
	if out.Source != SourceNPC || out.FallbackReason != FallbackReasonTimeout || out.FallbackDetail != "edge proxy timeout" {
		t.Fatalf("%+v", out)
	}
}

func TestInferResultToHTTPNPCFallbackJSON(t *testing.T) {
	in := InferResult{
		Text:           NPCFallback("scout"),
		Source:         SourceNPC,
		FallbackReason: FallbackReasonHTTPStatus,
		FallbackDetail: "502:backend_infer_failed",
		RAGK:           5,
		RAGHitIDs:      []string{"tactic-flank", "wei-cao"},
	}
	raw, err := json.Marshal(InferResultToHTTP(in, "", 0))
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatal(err)
	}
	if m["source"] != SourceNPC || m["fallback_reason"] != string(FallbackReasonHTTPStatus) {
		t.Fatalf("%v", m)
	}
	if m["fallback_detail"] != "502:backend_infer_failed" {
		t.Fatalf("detail=%v", m["fallback_detail"])
	}
	if m["rag_k"] != float64(5) {
		t.Fatalf("rag_k=%v", m["rag_k"])
	}
	ids, ok := m["rag_hit_ids"].([]interface{})
	if !ok || len(ids) != 2 {
		t.Fatalf("rag_hit_ids=%v", m["rag_hit_ids"])
	}
}

func TestInferResultToHTTPRoundTripEdge(t *testing.T) {
	orig := InferResultFromHTTP(EdgeInferHTTPResponseWithRAG("ok", "m", 2, 3, []string{"a"}))
	back := InferResultFromHTTP(InferResultToHTTP(orig, "m", 2))
	if back.Text != orig.Text || back.Source != orig.Source || back.RAGK != orig.RAGK || len(back.RAGHitIDs) != 1 {
		t.Fatalf("orig=%+v back=%+v", orig, back)
	}
}
