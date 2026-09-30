package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
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
}

// InferResult is a successful model response.
type InferResult struct {
	Text string
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

// Infer posts a prompt; on error or timeout returns NPC template within RequestTimeout.
func (c *InferClient) Infer(ctx context.Context, persona, prompt string) InferResult {
	reqTimeout := c.requestTimeout()
	ctx, cancel := context.WithTimeout(ctx, reqTimeout)
	defer cancel()
	body, _ := json.Marshal(map[string]string{"prompt": prompt})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/v1/infer", bytes.NewReader(body))
	if err != nil {
		return InferResult{Text: NPCFallback(persona)}
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient().Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		if resp != nil {
			resp.Body.Close()
		}
		return InferResult{Text: NPCFallback(persona)}
	}
	defer resp.Body.Close()
	var decoded struct {
		Text string `json:"text"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil || decoded.Text == "" {
		return InferResult{Text: NPCFallback(persona)}
	}
	return InferResult{Text: decoded.Text}
}

// BaseURLMust panics when url empty (tests).
func BaseURLMust(base string) string {
	if base == "" {
		panic(fmt.Sprintf("empty infer base"))
	}
	return base
}
