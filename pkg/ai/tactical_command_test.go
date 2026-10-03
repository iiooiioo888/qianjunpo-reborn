package ai

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/iiooiioo888/qianjunpo-reborn/pkg/tactical"
)

func TestTacticalCommandFromSuggestionMove(t *testing.T) {
	sug := moveSuggestion(InferPathEdge, 5, 8)
	cmd := TacticalCommandFromSuggestion("default/0", sug)
	if cmd.Kind != uint32(tactical.KindMove) || cmd.UnitID != tactical.UnitIDPlayer0 {
		t.Fatalf("%+v", cmd)
	}
	if cmd.BattleID != "default/0" || cmd.ToX != 5 || cmd.ToY != 8 {
		t.Fatalf("%+v", cmd)
	}
	if cmd.SkillID != 0 {
		t.Fatalf("skill_id should be omitted for move: %+v", cmd)
	}
}

func TestTacticalCommandFromSuggestionSkill(t *testing.T) {
	sug := skillSuggestion(InferPathRAG, 9, 9)
	cmd := TacticalCommandFromSuggestion("b1", sug)
	if cmd.Kind != uint32(tactical.KindSkill) || cmd.SkillID != 1 {
		t.Fatalf("%+v", cmd)
	}
}

func TestRunStaggeredSuggestInferFallbackNonEmpty(t *testing.T) {
	ResetStaggerForTests()
	ForceStaggerPath(InferPathEdge)
	ctx := context.Background()
	inferFn := func(_ context.Context, _ string) (string, string, int64, error) {
		return "", "", 0, errors.New("backend unavailable")
	}
	out, err := RunStaggeredSuggest(ctx, nil, 0, inferFn, StaggeredSuggestInput{
		BattleID: "default/0",
		Persona:  "guard",
		Order:    "hold east",
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if out.InferHTTPResponse.Text == "" {
		t.Fatal("expected npc text")
	}
	if out.InferHTTPResponse.Source != SourceNPC {
		t.Fatalf("source=%s", out.InferHTTPResponse.Source)
	}
	if out.InferHTTPResponse.Suggestion == nil {
		t.Fatal("missing suggestion")
	}
	if out.InferHTTPResponse.Command == nil {
		t.Fatal("missing command")
	}
	if out.InferHTTPResponse.Command.Kind != uint32(tactical.KindMove) {
		t.Fatalf("command=%+v", out.InferHTTPResponse.Command)
	}
	raw, err := json.Marshal(out.InferHTTPResponse)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatal(err)
	}
	if m["command"] == nil || m["suggestion"] == nil {
		t.Fatalf("payload=%v", m)
	}
}

func TestAttachTacticalCommandOnSuccessPath(t *testing.T) {
	ResetStaggerForTests()
	ForceStaggerPath(InferPathEdge)
	ctx := context.Background()
	inferFn := func(_ context.Context, prompt string) (string, string, int64, error) {
		return "advance: " + prompt, "mock", 1, nil
	}
	out, err := RunStaggeredSuggest(ctx, nil, 0, inferFn, StaggeredSuggestInput{
		BattleID: "default/0",
		Persona:  "guard",
		Order:    "flank",
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if out.InferHTTPResponse.Command == nil || out.InferHTTPResponse.Command.BattleID != "default/0" {
		t.Fatalf("command=%+v", out.InferHTTPResponse.Command)
	}
}
