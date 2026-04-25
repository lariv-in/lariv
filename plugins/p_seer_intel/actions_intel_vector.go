package p_seer_intel

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/lariv-in/lago/plugins/p_google_genai"
	"github.com/pgvector/pgvector-go"
	"google.golang.org/genai"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// EmbedQueryText returns an embedding vector for arbitrary text using the same
// embedding model and dimension as [NewFromIntelKind].
func EmbedQueryText(ctx context.Context, text string) ([]float32, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil, fmt.Errorf("p_seer_intel: EmbedQueryText: empty text")
	}
	client, err := p_google_genai.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	model := strings.TrimSpace(IntelConfigValue.EmbeddingModel)
	if model == "" {
		return nil, fmt.Errorf("p_seer_intel: EmbedQueryText: embeddingModel is empty")
	}
	valuesResp, err := client.Models.EmbedContent(ctx,
		model,
		[]*genai.Content{genai.NewContentFromText(text, genai.RoleUser)},
		nil,
	)
	if err != nil {
		slog.Error("p_seer_intel: embed query", "error", err)
		return nil, fmt.Errorf("p_seer_intel: embed query: %w", err)
	}
	values := valuesResp.Embeddings[0].Values
	if len(values) != SeerIntelEmbeddingDim {
		return nil, fmt.Errorf("p_seer_intel: embed dimension %d, want %d", len(values), SeerIntelEmbeddingDim)
	}
	return values, nil
}

// SearchIntelBySimilarity returns up to [limit] intel rows ordered by pgvector cosine
// distance to the embedding of [query] (smaller distance = more similar).
// Requires PostgreSQL and rows with non-null [Intel.Embedding].
func SearchIntelBySimilarity(ctx context.Context, db *gorm.DB, query string, limit int) ([]Intel, error) {
	if db == nil {
		return nil, fmt.Errorf("p_seer_intel: SearchIntelBySimilarity: db is nil")
	}
	if db.Name() != "postgres" {
		return nil, fmt.Errorf("p_seer_intel: SearchIntelBySimilarity: only postgres is supported (got %q)", db.Name())
	}
	if limit <= 0 {
		limit = 10
	}

	values, err := EmbedQueryText(ctx, query)
	if err != nil {
		return nil, err
	}
	vec := pgvector.NewVector(values)

	var rows []Intel
	// Wrap clause.Expr in clause.OrderBy — db.Order(clause.Expr{...}) is a no-op in GORM.
	err = db.WithContext(ctx).Model(&Intel{}).
		Where("embedding IS NOT NULL").
		Order(clause.OrderBy{
			Expression: clause.Expr{SQL: "embedding <=> ? ASC", Vars: []any{vec}},
		}).
		Limit(limit).
		Find(&rows).Error
	if err != nil {
		slog.Error("p_seer_intel: vector search", "error", err)
		return nil, fmt.Errorf("p_seer_intel: vector search: %w", err)
	}
	return rows, nil
}
