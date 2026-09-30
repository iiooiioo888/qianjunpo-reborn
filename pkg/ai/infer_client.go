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
	FallbackReason FallbackReason // set when Source == SourceNPC
	FallbackDetail string         // e.g. HTTP status or backend error code
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
	if c.shouldLogFallback() {
		logNPCFallback(persona, reason, detail)
	}
	return npcResult(persona, reason, detail)
}

// Infer posts a prompt; on error or timeout returns NPC template within RequestTimeout.
func (c *InferClient) Infer(ctx context.Context, persona, prompt string) InferResult {
	reqTimeout := c.requestTimeout()
	ctx, cancel := context.WithTimeout(ctx, reqTimeout)
	defer cancel()
	body, _ := json.Marshal(map[string]string{"prompt": prompt})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/v1/infer", bytes.NewReader(body))
	if err != nil {
		return c.fallback(persona, FallbackReasonRequestBuild, err.Error())
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
		return c.fallback(persona, reason, err.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		detail := strconv.Itoa(resp.StatusCode)
		if code := readInferErrorCode(resp.Body); code != "" {
			detail = detail + ":" + code
		}
		return c.fallback(persona, FallbackReasonHTTPStatus, detail)
	}
	var decoded struct {
		Text string `json:"text"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil || decoded.Text == "" {
		detail := ""
		if err != nil {
			detail = err.Error()
		}
		return c.fallback(persona, FallbackReasonBadResponse, detail)
	}
	return edgeResult(decoded.Text)
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
