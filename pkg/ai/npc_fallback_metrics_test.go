package ai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func assertNPCFallbackCounterIncrements(t *testing.T, reason FallbackReason, run func()) {
	t.Helper()
	ResetNPCFallbackCounts()
	key := string(reason)
	before := NPCFallbackCounts()[key]
	run()
	after := NPCFallbackCounts()[key]
	if after <= before {
		t.Fatalf("counter for %q did not increase: before=%d after=%d", reason, before, after)
	}
}

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

func TestInferClientFallbackReasonRequestBuildIncrementsCounter(t *testing.T) {
	reason := FallbackReasonRequestBuild
	assertNPCFallbackCounterIncrements(t, reason, func() {
		c := inferClientNoLog(":", nil)
		out := c.Infer(context.Background(), "guard", "hold")
		if out.Source != SourceNPC || out.FallbackReason != reason {
			t.Fatalf("%+v", out)
		}
	})
}

func TestInferClientFallbackReasonHTTPTransportIncrementsCounter(t *testing.T) {
	reason := FallbackReasonHTTPTransport
	assertNPCFallbackCounterIncrements(t, reason, func() {
		c := inferClientNoLog("http://127.0.0.1:1", nil)
		out := c.Infer(context.Background(), "scout", "probe")
		if out.FallbackReason != reason {
			t.Fatalf("reason=%q want %q", out.FallbackReason, reason)
		}
	})
}

func TestInferClientFallbackReasonTimeoutIncrementsCounter(t *testing.T) {
	reason := FallbackReasonTimeout
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(300 * time.Millisecond)
		_ = json.NewEncoder(w).Encode(map[string]string{"text": "late"})
	}))
	defer srv.Close()

	assertNPCFallbackCounterIncrements(t, reason, func() {
		c := inferClientNoLog(srv.URL, func(c *InferClient) {
			c.RequestTimeout = 200 * time.Millisecond
			c.HTTPTimeout = 150 * time.Millisecond
		})
		out := c.Infer(context.Background(), "guard", "hold")
		if out.FallbackReason != reason {
			t.Fatalf("reason=%q", out.FallbackReason)
		}
	})
}

func TestInferClientFallbackReasonHTTPStatusIncrementsCounter(t *testing.T) {
	reason := FallbackReasonHTTPStatus
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_ = json.NewEncoder(w).Encode(map[string]string{"code": "backend_infer_failed"})
	}))
	defer srv.Close()

	assertNPCFallbackCounterIncrements(t, reason, func() {
		out := inferClientNoLog(srv.URL, nil).Infer(context.Background(), "scout", "north")
		if out.FallbackReason != reason {
			t.Fatalf("%+v", out)
		}
	})
}

func TestInferClientFallbackReasonBadResponseEmptyIncrementsCounter(t *testing.T) {
	reason := FallbackReasonBadResponse
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"text":""}`))
	}))
	defer srv.Close()

	assertNPCFallbackCounterIncrements(t, reason, func() {
		out := inferClientNoLog(srv.URL, nil).Infer(context.Background(), "strategist", "wait")
		if out.FallbackReason != reason {
			t.Fatalf("%+v", out)
		}
	})
}

func TestInferClientFallbackReasonBadResponseJSONIncrementsCounter(t *testing.T) {
	reason := FallbackReasonBadResponse
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`not-json`))
	}))
	defer srv.Close()

	assertNPCFallbackCounterIncrements(t, reason, func() {
		out := inferClientNoLog(srv.URL, nil).Infer(context.Background(), "merchant", "move")
		if out.FallbackReason != reason {
			t.Fatalf("%+v", out)
		}
	})
}

func TestInferClientProxyNPCOnWireIncrementsCounter(t *testing.T) {
	reason := FallbackReasonTimeout
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(InferHTTPResponse{
			Text:           "edge gave up",
			Source:         SourceNPC,
			FallbackReason: string(reason),
			FallbackDetail: "edge proxy timeout",
		})
	}))
	defer srv.Close()

	assertNPCFallbackCounterIncrements(t, reason, func() {
		out := inferClientNoLog(srv.URL, nil).Infer(context.Background(), "guard", "hold")
		if out.FallbackReason != reason || out.Text != NPCFallback("guard") {
			t.Fatalf("%+v", out)
		}
	})
}

func TestRAGInferTimeoutIncrementsCounter(t *testing.T) {
	reason := FallbackReasonTimeout
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(300 * time.Millisecond)
	}))
	defer srv.Close()

	assertNPCFallbackCounterIncrements(t, reason, func() {
		client, err := DefaultRAGInferClient(srv.URL)
		if err != nil {
			t.Fatal(err)
		}
		client.Infer.RequestTimeout = 250 * time.Millisecond
		client.Infer.HTTPTimeout = 200 * time.Millisecond
		off := false
		client.Infer.LogFallback = &off
		out := client.InferWithContext(context.Background(), "guard", "flank")
		if out.FallbackReason != reason {
			t.Fatalf("%+v", out)
		}
	})
}
