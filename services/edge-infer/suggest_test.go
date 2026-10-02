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
