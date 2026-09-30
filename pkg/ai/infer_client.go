package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// InferClient calls the edge inference HTTP API with fast fallback.
type InferClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

// InferResult is a successful model response.
type InferResult struct {
	Text string
}

// Infer posts a prompt; on error or timeout returns NPC template within 1s.
func (c *InferClient) Infer(ctx context.Context, persona, prompt string) InferResult {
	if c.HTTPClient == nil {
		c.HTTPClient = &http.Client{Timeout: 800 * time.Millisecond}
	}
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	body, _ := json.Marshal(map[string]string{"prompt": prompt})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/v1/infer", bytes.NewReader(body))
	if err != nil {
		return InferResult{Text: NPCFallback(persona)}
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.HTTPClient.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
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
