package ai

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/iiooiioo888/qianjunpo-reborn/pkg/rag"
)

// RAGInferClient retrieves Top-K lore then calls edge inference.
// Retrieval + HTTP are outside FP64 combat paths.
type RAGInferClient struct {
	Infer   *InferClient
	Store   rag.Store
	TopK    int
	// LogRAG emits structured slog when Top-K hits augment the infer prompt; default true.
	LogRAG *bool
}

// InferWithContext runs RAG retrieve → augmented prompt → InferClient (NPC fallback <1s on failure).
func (c *RAGInferClient) InferWithContext(ctx context.Context, persona, userPrompt string) InferResult {
	k := c.TopK
	if k <= 0 {
		k = rag.DefaultTopK
	}
	augmented := userPrompt
	var hitIDs []string
	if c.Store != nil {
		hits, err := c.Store.TopKText(ctx, userPrompt, k)
		if err == nil && len(hits) > 0 {
			hitIDs = rag.DocumentIDs(hits)
			augmented = buildRAGPrompt(rag.FormatContext(hits), persona, userPrompt)
			if c.shouldLogRAG() {
				logRAGAugmented(persona, k, hitIDs)
			}
		}
	}
	if c.Infer == nil {
		out := npcFallbackResult(persona, FallbackReasonNoInferClient, "", c.shouldLogFallback())
		return withRAGMeta(out, k, hitIDs)
	}
	return c.Infer.InferWithRAGMeta(ctx, persona, augmented, k, hitIDs)
}

func (c *RAGInferClient) shouldLogFallback() bool {
	if c.Infer != nil {
		return c.Infer.shouldLogFallback()
	}
	return true
}

func (c *RAGInferClient) shouldLogRAG() bool {
	if c.LogRAG != nil {
		return *c.LogRAG
	}
	return true
}

func logRAGAugmented(persona string, ragK int, hitIDs []string) {
	attrs := []any{
		slog.Int("rag_k", ragK),
		slog.Any("rag_hit_ids", hitIDs),
	}
	if persona != "" {
		attrs = append(attrs, slog.String("persona", persona))
	}
	slog.Default().Info("ai infer rag augmented", attrs...)
}

func buildRAGPrompt(contextBlock, persona, userPrompt string) string {
	var b strings.Builder
	b.WriteString("You are a Three Kingdoms tactical advisor")
	if persona != "" {
		b.WriteString(" (")
		b.WriteString(persona)
		b.WriteString(")")
	}
	b.WriteString(".\nKnowledge:\n")
	b.WriteString(contextBlock)
	b.WriteString("\n\nOrder: ")
	b.WriteString(userPrompt)
	b.WriteString("\nReply with one short tactical line.")
	return b.String()
}

// DefaultRAGInferClient wires a seeded in-memory store for local dev/tests.
func DefaultRAGInferClient(baseURL string) (*RAGInferClient, error) {
	store, err := rag.SeedStore(context.Background(), rag.NewHashBagEmbedder(64))
	if err != nil {
		return nil, fmt.Errorf("seed rag: %w", err)
	}
	return &RAGInferClient{
		Infer: &InferClient{BaseURL: baseURL},
		Store: store,
		TopK:  rag.DefaultTopK,
	}, nil
}
