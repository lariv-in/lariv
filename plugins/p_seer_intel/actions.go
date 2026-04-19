package p_seer_intel

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/pgvector/pgvector-go"
	"google.golang.org/genai"
)

const intelSummarySystemPrompt = `You write concise factual summaries for an intelligence ingest pipeline.
Given raw source content, respond with a short plain-text summary only (no markdown headings, no preamble).
Aim for 2–6 sentences. If the content is empty or unusable, reply with a single sentence stating that.`

const intelTitleSystemPrompt = `You label rows in an intelligence ingest pipeline.
Given raw source content, respond with one short plain-text title only: no markdown, no preamble, no quotation marks.
At most 12 words. Describe the subject (what it is about), not the medium (avoid "post", "article", "document" unless necessary).`

const intelTitleMaxRunes = 200

// normalizeIntelTitle cleans model output to a single-line DB title.
func normalizeIntelTitle(raw string) string {
	s := strings.TrimSpace(raw)
	if s == "" {
		return ""
	}
	if i := strings.IndexAny(s, "\r\n"); i >= 0 {
		s = strings.TrimSpace(s[:i])
	}
	s = strings.Trim(s, `"'`)
	if utf8.RuneCountInString(s) <= intelTitleMaxRunes {
		return s
	}
	runes := []rune(s)
	return strings.TrimSpace(string(runes[:intelTitleMaxRunes]))
}

func intelTitleFallback(content string) string {
	for _, line := range strings.Split(content, "\n") {
		t := strings.TrimSpace(line)
		if t == "" || strings.HasPrefix(t, "---") {
			continue
		}
		t = normalizeIntelTitle(t)
		if t != "" {
			return t
		}
	}
	return "Intel"
}

// NewFromIntelKind builds an [Intel] from k using Google Gen AI: the LLM produces Title and Summary from k.Content(),
// the embedding model produces [Intel.Embedding] at [SeerIntelEmbeddingDim], and Datetime is UTC now.
// [Intel.Kind] / [Intel.KindID] are set from k.
//
// Requires [IntelGenAI.APIKey] from app config ([Plugins] "p_seer_intel".apiKey). Returns an error if the key is missing,
// content is empty, or either API call fails.
func NewFromIntelKind(ctx context.Context, k IntelKind) (Intel, error) {
	if k == nil {
		return Intel{}, nil
	}
	content := k.Content()
	if strings.TrimSpace(content) == "" {
		return Intel{}, fmt.Errorf("p_seer_intel: NewFromIntelKind: empty content")
	}
	key := strings.TrimSpace(IntelGenAI.APIKey)
	if key == "" {
		return Intel{}, fmt.Errorf("p_seer_intel: NewFromIntelKind: Plugins.p_seer_intel apiKey is empty")
	}

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  key,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return Intel{}, fmt.Errorf("p_seer_intel: genai client: %w", err)
	}

	llmModel := strings.TrimSpace(IntelGenAI.LLMModel)
	if llmModel == "" {
		llmModel = defaultIntelLLMModel
	}
	embedModel := strings.TrimSpace(IntelGenAI.EmbeddingModel)
	if embedModel == "" {
		embedModel = defaultIntelEmbeddingModel
	}

	titleCfg := &genai.GenerateContentConfig{
		Temperature:       genai.Ptr[float32](0.2),
		SystemInstruction: genai.NewContentFromText(intelTitleSystemPrompt, genai.RoleUser),
		MaxOutputTokens:   128,
	}
	titleResp, err := client.Models.GenerateContent(ctx, llmModel, genai.Text(content), titleCfg)
	if err != nil {
		slog.Error("p_seer_intel: title generate", "error", err, "model", llmModel)
		return Intel{}, fmt.Errorf("p_seer_intel: title generate: %w", err)
	}
	title := normalizeIntelTitle(titleResp.Text())
	if title == "" {
		title = intelTitleFallback(content)
	}

	sumCfg := &genai.GenerateContentConfig{
		Temperature:       genai.Ptr[float32](0.3),
		SystemInstruction: genai.NewContentFromText(intelSummarySystemPrompt, genai.RoleUser),
	}
	sumResp, err := client.Models.GenerateContent(ctx, llmModel, genai.Text(content), sumCfg)
	if err != nil {
		slog.Error("p_seer_intel: summary generate", "error", err, "model", llmModel)
		return Intel{}, fmt.Errorf("p_seer_intel: summary generate: %w", err)
	}
	summary := strings.TrimSpace(sumResp.Text())
	if summary == "" {
		return Intel{}, fmt.Errorf("p_seer_intel: summary generate returned empty text")
	}

	dim := int32(SeerIntelEmbeddingDim)
	embedCfg := &genai.EmbedContentConfig{OutputDimensionality: &dim}
	embedContents := []*genai.Content{genai.NewContentFromText(content, genai.RoleUser)}
	embedRes, err := client.Models.EmbedContent(ctx, embedModel, embedContents, embedCfg)
	if err != nil {
		slog.Error("p_seer_intel: embed content", "error", err, "model", embedModel)
		return Intel{}, fmt.Errorf("p_seer_intel: embed content: %w", err)
	}
	if len(embedRes.Embeddings) == 0 || embedRes.Embeddings[0] == nil {
		return Intel{}, fmt.Errorf("p_seer_intel: embed content returned no embeddings")
	}
	values := embedRes.Embeddings[0].Values
	if len(values) != SeerIntelEmbeddingDim {
		return Intel{}, fmt.Errorf("p_seer_intel: embed dimension %d, want %d", len(values), SeerIntelEmbeddingDim)
	}

	vec := pgvector.NewVector(values)
	return Intel{
		Title:     title,
		Kind:      strings.TrimSpace(k.Kind()),
		KindID:    k.IntelID(),
		Summary:   summary,
		Datetime:  time.Now().UTC(),
		Embedding: &vec,
	}, nil
}
