package ai

import (
	"context"
	"testing"
)

func TestInferClientNPCFallbackIncrementsCounter(t *testing.T) {
	ResetNPCFallbackCounts()
	off := false
	c := &InferClient{BaseURL: "http://127.0.0.1:1", LogFallback: &off}
	before := NPCFallbackCounts()
	out := c.Infer(context.Background(), "guard", "defend")
	if out.Source != SourceNPC {
		t.Fatalf("expected npc source, got %+v", out)
	}
	after := NPCFallbackCounts()
	if after[string(out.FallbackReason)] <= before[string(out.FallbackReason)] {
		t.Fatalf("counter for %q did not increase: before=%v after=%v", out.FallbackReason, before, after)
	}
}
