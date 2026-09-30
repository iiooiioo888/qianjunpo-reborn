package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func withSlogCapture(t *testing.T, fn func()) string {
	var buf bytes.Buffer
	prev := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo})))
	defer slog.SetDefault(prev)
	fn()
	return buf.String()
}

func TestInferClientLogsNPCFallbackFields(t *testing.T) {
	logs := withSlogCapture(t, func() {
		c := &InferClient{BaseURL: "http://127.0.0.1:1"}
		out := c.Infer(context.Background(), "scout", "probe")
		if out.Source != SourceNPC {
			t.Fatalf("%+v", out)
		}
	})
	if !strings.Contains(logs, "ai infer npc fallback") {
		t.Fatalf("missing log line: %q", logs)
	}
	var entry map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(logs)), &entry); err != nil {
		t.Fatalf("parse log: %v raw=%q", err, logs)
	}
	if entry["source"] != SourceNPC {
		t.Fatalf("source=%v", entry["source"])
	}
	reason, _ := entry["fallback_reason"].(string)
	if reason != string(FallbackReasonHTTPTransport) && reason != string(FallbackReasonTimeout) {
		t.Fatalf("fallback_reason=%v", entry["fallback_reason"])
	}
	if entry["fallback_detail"] == nil || entry["fallback_detail"] == "" {
		t.Fatalf("expected fallback_detail, got %v", entry["fallback_detail"])
	}
}

func TestInferClientSkipsNPCFallbackLogWhenDisabled(t *testing.T) {
	off := false
	logs := withSlogCapture(t, func() {
		c := &InferClient{BaseURL: "http://127.0.0.1:1", LogFallback: &off}
		_ = c.Infer(context.Background(), "guard", "hold")
	})
	if strings.TrimSpace(logs) != "" {
		t.Fatalf("expected no logs, got %q", logs)
	}
}

func TestInferClientLogsHTTPStatusFallbackDetail(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_ = json.NewEncoder(w).Encode(map[string]string{"code": "backend_infer_failed"})
	}))
	defer srv.Close()

	logs := withSlogCapture(t, func() {
		c := &InferClient{BaseURL: srv.URL}
		out := c.Infer(context.Background(), "scout", "north")
		if out.FallbackReason != FallbackReasonHTTPStatus {
			t.Fatalf("%+v", out)
		}
	})
	var entry map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(logs)), &entry); err != nil {
		t.Fatal(err)
	}
	detail, _ := entry["fallback_detail"].(string)
	if !strings.Contains(detail, "502") || !strings.Contains(detail, "backend_infer_failed") {
		t.Fatalf("fallback_detail=%q", detail)
	}
}
