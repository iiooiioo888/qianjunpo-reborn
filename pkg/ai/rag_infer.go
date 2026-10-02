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

// RAGRetrieveOutcome is the augmented prompt and observability from Top-K retrieval.
type RAGRetrieveOutcome struct {
	AugmentedPrompt string
	K               int
	HitIDs          []string
}

// RetrieveAndAugment runs Top-K using battle context + order, then builds the infer prompt.
func RetrieveAndAugment(ctx context.Context, store rag.Store, topK int, persona, order string, battle BattleContext) RAGRetrieveOutcome {
	k := topK
	if k <= 0 {
		k = rag.DefaultTopK
	}
	out := RAGRetrieveOutcome{K: k, AugmentedPrompt: order}
	if store == nil {
		if battleBlock := battle.PromptBlock(); battleBlock != "" {
			out.AugmentedPrompt = buildRAGPrompt("", battleBlock, persona, order)
		}
		return out
	}
	query := battle.RetrievalQuery(order)
	if query == "" {
		query = order
	}
	hits, err := store.TopKText(ctx, query, k)
	if err != nil || len(hits) == 0 {
		if battleBlock := battle.PromptBlock(); battleBlock != "" {
			out.AugmentedPrompt = buildRAGPrompt("", battleBlock, persona, order)
		}
		return out
	}
	out.HitIDs = rag.DocumentIDs(hits)
	contextBlock := rag.FormatContext(hits)
	battleBlock := battle.PromptBlock()
	out.AugmentedPrompt = buildRAGPrompt(contextBlock, battleBlock, persona, order)
	return out
}

// InferWithContext runs RAG retrieve → augmented prompt → InferClient (NPC fallback <1s on failure).
func (c *RAGInferClient) InferWithContext(ctx context.Context, persona, userPrompt string) InferResult {
	return c.InferWithBattle(ctx, persona, userPrompt, BattleContext{})
}

// InferWithBattle includes live unit/terrain summary in retrieval and prompt.
func (c *RAGInferClient) InferWithBattle(ctx context.Context, persona, userPrompt string, battle BattleContext) InferResult {
	k := c.TopK
	if k <= 0 {
		k = rag.DefaultTopK
	}
	retrieved := RetrieveAndAugment(ctx, c.Store, k, persona, userPrompt, battle)
	if len(retrieved.HitIDs) > 0 && c.shouldLogRAG() {
		logRAGAugmented(persona, retrieved.K, retrieved.HitIDs)
	}
	if c.Infer == nil {
		out := npcFallbackResult(persona, FallbackReasonNoInferClient, "", c.shouldLogFallback())
		return withRAGMeta(out, retrieved.K, retrieved.HitIDs)
	}
	return c.Infer.InferWithRAGMeta(ctx, persona, retrieved.AugmentedPrompt, retrieved.K, retrieved.HitIDs)
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

func buildRAGPrompt(contextBlock, battleBlock, persona, userPrompt string) string {
	var b strings.Builder
	b.WriteString("You are a Three Kingdoms tactical advisor")
	if persona != "" {
		b.WriteString(" (")
		b.WriteString(persona)
		b.WriteString(")")
	}
	b.WriteString(".\n")
	if battleBlock != "" {
		b.WriteString(battleBlock)
		b.WriteString("\n")
	}
	if contextBlock != "" {
		b.WriteString("Knowledge:\n")
		b.WriteString(contextBlock)
		b.WriteString("\n")
	}
	b.WriteString("\nOrder: ")
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
