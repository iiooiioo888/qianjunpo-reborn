package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/iiooiioo888/qianjunpo-reborn/pkg/ai"
	"github.com/iiooiioo888/qianjunpo-reborn/services/edge-infer/backend"
)

func TestSuggestHTTPWriteBackAndGet(t *testing.T) {
	ai.ResetStaggerForTests()
	ai.GlobalSuggestionStore().ResetSuggestions()
	m := &backend.MockBackend{}
	srv := &server{backend: m, mock: m}

	body := bytes.NewBufferString(`{
		"battle_id":"default/0",
		"persona":"guard",
		"order":"flank east with cavalry",
		"battle_context":{"units_summary":"cavalry wing","terrain_summary":"hills"}
	}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/suggest", body)
	rr := httptest.NewRecorder()
	srv.handleSuggest(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("post code=%d body=%s", rr.Code, rr.Body.String())
	}
	var postResp ai.InferHTTPResponse
	if err := json.NewDecoder(rr.Body).Decode(&postResp); err != nil {
		t.Fatal(err)
	}
	if postResp.Suggestion == nil {
		t.Fatalf("expected suggestion in POST response: %+v", postResp)
	}
	if postResp.Suggestion.Kind != ai.SuggestionKindMove && postResp.Suggestion.Kind != ai.SuggestionKindSkill {
		t.Fatalf("suggestion=%+v", postResp.Suggestion)
	}
	if postResp.Command == nil || postResp.Command.UnitID == 0 {
		t.Fatalf("command=%+v", postResp.Command)
	}

	getReq := httptest.NewRequest(http.MethodGet, "/v1/suggest/write-back?battle_id=default/0", nil)
	getRR := httptest.NewRecorder()
	srv.handleSuggestGet(getRR, getReq)
	if getRR.Code != http.StatusOK {
		t.Fatalf("get code=%d body=%s", getRR.Code, getRR.Body.String())
	}
	var getPayload map[string]interface{}
	if err := json.NewDecoder(getRR.Body).Decode(&getPayload); err != nil {
		t.Fatal(err)
	}
	sug, ok := getPayload["suggestion"].(map[string]interface{})
	if !ok || sug["kind"] == nil {
		t.Fatalf("payload=%v", getPayload)
	}
	cmd, ok := getPayload["command"].(map[string]interface{})
	if !ok || cmd["kind"] == nil {
		t.Fatalf("payload=%v", getPayload)
	}
}

func TestSuggestHTTPInferDownFallback(t *testing.T) {
	ai.ResetStaggerForTests()
	ai.GlobalSuggestionStore().ResetSuggestions()
	srv := &server{backend: inferFailBackend{}}
	body := bytes.NewBufferString(`{
		"battle_id":"default/0",
		"persona":"guard",
		"order":"hold the gate"
	}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/suggest", body)
	rr := httptest.NewRecorder()
	srv.handleSuggest(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("post code=%d body=%s", rr.Code, rr.Body.String())
	}
	var postResp ai.InferHTTPResponse
	if err := json.NewDecoder(rr.Body).Decode(&postResp); err != nil {
		t.Fatal(err)
	}
	if postResp.Source != ai.SourceNPC || postResp.Text == "" {
		t.Fatalf("expected npc fallback: %+v", postResp)
	}
	if postResp.Suggestion == nil || postResp.Command == nil {
		t.Fatalf("missing consumable fields: %+v", postResp)
	}
}

func TestSuggestDemoPage(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/v1/suggest/demo", nil)
	rr := httptest.NewRecorder()
	srv := &server{}
	srv.handleSuggestDemo(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("code=%d", rr.Code)
	}
	if ct := rr.Header().Get("Content-Type"); ct != "text/html; charset=utf-8" {
		t.Fatalf("content-type=%q", ct)
	}
	body := rr.Body.String()
	if !bytes.Contains(rr.Body.Bytes(), []byte("POST /v1/suggest")) {
		t.Fatalf("missing UI hook in demo HTML")
	}
	if body == "" {
		t.Fatal("empty demo page")
	}
}

func TestSuggestStaggerPathsAlternate(t *testing.T) {
	ai.ResetStaggerForTests()
	ai.GlobalSuggestionStore().ResetSuggestions()
	m := &backend.MockBackend{}
	srv := &server{backend: m, mock: m}
	post := func() string {
		body := bytes.NewBufferString(`{"persona":"guard","order":"hold line"}`)
		req := httptest.NewRequest(http.MethodPost, "/v1/suggest", body)
		rr := httptest.NewRecorder()
		srv.handleSuggest(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("code=%d", rr.Code)
		}
		var resp ai.InferHTTPResponse
		_ = json.NewDecoder(rr.Body).Decode(&resp)
		if resp.Suggestion == nil {
			t.Fatal("missing suggestion")
		}
		return resp.Suggestion.InferPath
	}
	p1 := post()
	p2 := post()
	if p1 == p2 {
		t.Fatalf("expected staggered paths, got %s %s", p1, p2)
	}
}
