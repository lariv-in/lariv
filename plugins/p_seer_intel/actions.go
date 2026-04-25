package p_seer_intel

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/lariv-in/lago/plugins/p_google_genai"
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

// NewFromIntelKind builds an [Intel] from k using text + embeddings from [p_google_genai].inteaction
// [Intel.Kind] / [Intel.KindID] are set from k.
func NewFromIntelKind(ctx context.Context, k IntelKind) (Intel, error) {
	if k == nil {
		return Intel{}, nil
	}
	content := k.Content()
	if strings.TrimSpace(content) == "" {
		return Intel{}, fmt.Errorf("p_seer_intel: NewFromIntelKind: empty content")
	}
	client, err := p_google_genai.NewClient(ctx)
	if err != nil {
		return Intel{}, err
	}
	resp, err := client.Models.GenerateContent(ctx,
		IntelConfigValue.TitleModel,
		[]*genai.Content{genai.NewContentFromText(content, genai.RoleUser)},
		&genai.GenerateContentConfig{
			SystemInstruction: genai.NewContentFromText(intelTitleSystemPrompt, genai.RoleUser),
		})
	if err != nil {
		slog.Error("p_seer_intel: title generate", "error", err)
		return Intel{}, fmt.Errorf("p_seer_intel: title generate: %w", err)
	}
	title := normalizeIntelTitle(resp.Text())
	if title == "" {
		title = intelTitleFallback(content)
	}

	summaryResp, err := client.Models.GenerateContent(ctx,
		IntelConfigValue.SummaryModel,
		[]*genai.Content{genai.NewContentFromText(content, genai.RoleUser)},
		&genai.GenerateContentConfig{
			SystemInstruction: genai.NewContentFromText(intelSummarySystemPrompt, genai.RoleUser),
		})
	if err != nil {
		slog.Error("p_seer_intel: summary generate", "error", err)
		return Intel{}, fmt.Errorf("p_seer_intel: summary generate: %w", err)
	}
	summary := strings.TrimSpace(summaryResp.Text())
	if summary == "" {
		return Intel{}, fmt.Errorf("p_seer_intel: summary generate returned empty text")
	}

	valuesResp, err := client.Models.EmbedContent(ctx,
		IntelConfigValue.EmbeddingModel,
		[]*genai.Content{genai.NewContentFromText(content, genai.RoleUser)},
		nil,
	)
	if err != nil {
		slog.Error("p_seer_intel: embed content", "error", err)
		return Intel{}, fmt.Errorf("p_seer_intel: embed content: %w", err)
	}
	values := valuesResp.Embeddings[0].Values
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
