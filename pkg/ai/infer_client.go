package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

// DefaultInferRequestTimeout bounds the full Infer call (RAG + HTTP); NPC fallback must finish within this window.
const DefaultInferRequestTimeout = time.Second

// DefaultInferHTTPTimeout is the per-request HTTP client limit (slightly under RequestTimeout).
const DefaultInferHTTPTimeout = 800 * time.Millisecond

// InferClient calls the edge inference HTTP API with fast fallback.
type InferClient struct {
	BaseURL string
	// HTTPClient overrides the client used for POST /v1/infer. When nil, one is built from HTTPTimeout.
	HTTPClient *http.Client
	// RequestTimeout caps the entire Infer call; zero uses DefaultInferRequestTimeout.
	RequestTimeout time.Duration
	// HTTPTimeout caps the outbound HTTP round-trip; zero uses DefaultInferHTTPTimeout.
	HTTPTimeout time.Duration
	// LogFallback emits structured slog when NPC fallback is used; default true.
	LogFallback *bool
}

// InferResult is edge model text or NPC fallback with observability fields.
type InferResult struct {
	Text           string
	Source         string         // SourceEdge or SourceNPC
	RAGK           int            // Top-K depth used when RAG augmented the prompt (0 when none)
	RAGHitIDs      []string       // retrieved chunk ids in rank order when RAG ran
	FallbackReason FallbackReason // set when Source == SourceNPC
	FallbackDetail string         // e.g. HTTP status or backend error code
}

// WithRAGObservability attaches RAG Top-K metadata (safe for nil hit slice).
func (r InferResult) WithRAGObservability(k int, hitIDs []string) InferResult {
	return withRAGMeta(r, k, hitIDs)
}

func (c *InferClient) requestTimeout() time.Duration {
	if c.RequestTimeout > 0 {
		return c.RequestTimeout
	}
	if v := os.Getenv("AI_INFER_REQUEST_TIMEOUT"); v != "" {
		if d, err := time.ParseDuration(v); err == nil && d > 0 {
			return d
		}
	}
	return DefaultInferRequestTimeout
}

func (c *InferClient) httpTimeout() time.Duration {
	if c.HTTPTimeout > 0 {
		return c.HTTPTimeout
	}
	if v := os.Getenv("AI_INFER_HTTP_TIMEOUT"); v != "" {
		if d, err := time.ParseDuration(v); err == nil && d > 0 {
			return d
		}
	}
	return DefaultInferHTTPTimeout
}

func (c *InferClient) httpClient() *http.Client {
	if c.HTTPClient != nil {
		return c.HTTPClient
	}
	return &http.Client{Timeout: c.httpTimeout()}
}

func (c *InferClient) shouldLogFallback() bool {
	if c.LogFallback != nil {
		return *c.LogFallback
	}
	return true
}

func (c *InferClient) fallback(persona string, reason FallbackReason, detail string) InferResult {
	return npcFallbackResult(persona, reason, detail, c.shouldLogFallback())
}

// Infer posts a prompt; on error or timeout returns NPC template within RequestTimeout.
func (c *InferClient) Infer(ctx context.Context, persona, prompt string) InferResult {
	return c.InferWithRAGMeta(ctx, persona, prompt, 0, nil)
}

// InferWithRAGMeta posts prompt plus optional RAG Top-K metadata for edge observability.
func (c *InferClient) InferWithRAGMeta(ctx context.Context, persona, prompt string, ragK int, ragHitIDs []string) InferResult {
	reqTimeout := c.requestTimeout()
	ctx, cancel := context.WithTimeout(ctx, reqTimeout)
	defer cancel()
	body, _ := json.Marshal(InferPostBody{Prompt: prompt, RAGK: ragK, RAGHitIDs: ragHitIDs})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/v1/infer", bytes.NewReader(body))
	if err != nil {
		return withRAGMeta(c.fallback(persona, FallbackReasonRequestBuild, err.Error()), ragK, ragHitIDs)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient().Do(req)
	if err != nil {
		reason := FallbackReasonHTTPTransport
		if errors.Is(err, context.DeadlineExceeded) {
			reason = FallbackReasonTimeout
		} else {
			var netErr interface{ Timeout() bool }
			if errors.As(err, &netErr) && netErr.Timeout() {
				reason = FallbackReasonTimeout
			}
		}
		return withRAGMeta(c.fallback(persona, reason, err.Error()), ragK, ragHitIDs)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		detail := strconv.Itoa(resp.StatusCode)
		if code := readInferErrorCode(resp.Body); code != "" {
			detail = detail + ":" + code
		}
		return withRAGMeta(c.fallback(persona, FallbackReasonHTTPStatus, detail), ragK, ragHitIDs)
	}
	var decoded InferHTTPResponse
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return withRAGMeta(c.fallback(persona, FallbackReasonBadResponse, err.Error()), ragK, ragHitIDs)
	}
	out := InferResultFromHTTP(decoded)
	if out.Text == "" {
		return withRAGMeta(c.fallback(persona, FallbackReasonBadResponse, ""), ragK, ragHitIDs)
	}
	if out.Source == SourceNPC {
		reason := out.FallbackReason
		if reason == FallbackReasonNone {
			reason = FallbackReasonBadResponse
		}
		return withRAGMeta(c.fallback(persona, reason, out.FallbackDetail), ragK, ragHitIDs)
	}
	return withRAGMeta(out, ragK, ragHitIDs)
}

func withRAGMeta(r InferResult, ragK int, ragHitIDs []string) InferResult {
	if ragK <= 0 && len(ragHitIDs) == 0 {
		return r
	}
	if r.RAGK == 0 {
		r.RAGK = ragK
	}
	if len(r.RAGHitIDs) == 0 && len(ragHitIDs) > 0 {
		r.RAGHitIDs = append([]string(nil), ragHitIDs...)
	}
	return r
}

type inferErrorBody struct {
	Code  string `json:"code"`
	Error string `json:"error"`
}

func readInferErrorCode(r io.Reader) string {
	var body inferErrorBody
	if err := json.NewDecoder(r).Decode(&body); err != nil {
		return ""
	}
	return body.Code
}

// BaseURLMust panics when url empty (tests).
func BaseURLMust(base string) string {
	if base == "" {
		panic(fmt.Sprintf("empty infer base"))
	}
	return base
}
