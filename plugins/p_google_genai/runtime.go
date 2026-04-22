// Package p_google_genai provides Gemini API / Vertex text generation and embeddings via
// [google.golang.org/genai].
package p_google_genai

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"strings"
	"sync"
	"time"

	"google.golang.org/genai"
)

// GenerateRequest holds parameters for text generation and embedding calls.
type GenerateRequest struct {
	SystemPrompt    string
	UserPrompt      string
	Temperature     *float32
	MaxOutputTokens int
	Thinking        *ThinkingConfig
}

// ThinkingConfig selects extended reasoning for models that support it.
type ThinkingConfig struct {
	Mode   string
	Budget int
}

// EmbedTask labels embedding inputs for task-type hints sent to the API.
type EmbedTask string

const (
	EmbedTaskSearchQuery    EmbedTask = "search_query"
	EmbedTaskSearchDocument EmbedTask = "search_document"
	EmbedTaskVisionDocument EmbedTask = "vision_document"
)

var (
	clientMu          sync.Mutex
	cachedClient      *genai.Client
	embedDim          int
	embedDimMu        sync.Mutex
	warnedEmptyAPIKey sync.Once
)

func GenerateText(ctx context.Context, req GenerateRequest) (string, error) {
	return GenerateTextStream(ctx, req, nil)
}

func GenerateTextStream(ctx context.Context, req GenerateRequest, onToken func(string) error) (string, error) {
	if strings.TrimSpace(req.UserPrompt) == "" {
		return "", fmt.Errorf("p_google_genai: GenerateTextStream: empty user prompt")
	}
	cli, err := genaiClientFor(ctx)
	if err != nil {
		return "", err
	}
	cfg := baseGenerateConfig(req)
	cfg.SystemInstruction = systemInstruction(req.SystemPrompt, req.Thinking)
	contents := genai.Text(strings.TrimSpace(req.UserPrompt))
	return runGenerateStream(ctx, cli, GoogleGenAIConfig.TextModel, contents, cfg, onToken)
}

func GenerateJSON(ctx context.Context, req GenerateRequest, target any) (string, error) {
	raw, err := GenerateText(ctx, req)
	if err != nil {
		return "", err
	}
	if err := decodeJSONBlob(raw, target); err == nil {
		return extractJSONBlob(raw), nil
	}
	last := raw
	for attempt := 0; attempt < 2; attempt++ {
		repairPrompt := fmt.Sprintf(
			"Return only valid JSON. Do not add markdown, commentary, or code fences.\n\nInvalid input:\n%s",
			last,
		)
		repaired, err := GenerateText(ctx, GenerateRequest{
			SystemPrompt:    "You repair invalid JSON. Keep the same meaning and output only valid JSON.",
			UserPrompt:      repairPrompt,
			MaxOutputTokens: max(req.MaxOutputTokens, 512),
			Thinking:        &ThinkingConfig{Mode: ThinkingModeDisabled},
		})
		if err != nil {
			return "", err
		}
		if err := decodeJSONBlob(repaired, target); err == nil {
			return extractJSONBlob(repaired), nil
		}
		last = repaired
	}
	return "", fmt.Errorf("p_google_genai: GenerateJSON: unable to parse model output")
}

func EmbedText(ctx context.Context, task EmbedTask, text string) ([]float32, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil, fmt.Errorf("p_google_genai: EmbedText: empty text")
	}
	cli, err := genaiClientFor(ctx)
	if err != nil {
		return nil, err
	}
	input := embeddingTaskPrefix(task) + text
	tt := embedTaskType(task)
	logEmbedRequest(ctx, task, input, tt)
	resp, err := withGenAIRetryResp(ctx, "embed_content", func() (*genai.EmbedContentResponse, error) {
		return cli.Models.EmbedContent(ctx, GoogleGenAIConfig.EmbeddingModel, genai.Text(input), &genai.EmbedContentConfig{
			TaskType: tt,
		})
	})
	if err != nil {
		slog.WarnContext(ctx, "p_google_genai: EmbedContent error", "model", GoogleGenAIConfig.EmbeddingModel, "error", err.Error())
		return nil, err
	}
	if resp == nil || len(resp.Embeddings) == 0 || resp.Embeddings[0] == nil {
		return nil, fmt.Errorf("p_google_genai: EmbedText: empty embedding response")
	}
	values := resp.Embeddings[0].Values
	if len(values) == 0 {
		return nil, fmt.Errorf("p_google_genai: EmbedText: model returned no values")
	}
	embedDimMu.Lock()
	if embedDim == 0 {
		embedDim = len(values)
	}
	embedDimMu.Unlock()
	return normalizeVector(values), nil
}

func EmbeddingDimension(ctx context.Context) (int, error) {
	if d := GoogleGenAIConfig.EmbeddingDimHint; d > 0 {
		return d, nil
	}
	embedDimMu.Lock()
	d := embedDim
	embedDimMu.Unlock()
	if d > 0 {
		return d, nil
	}
	if _, err := EmbedText(ctx, EmbedTaskSearchDocument, "dimension probe"); err != nil {
		return 0, err
	}
	embedDimMu.Lock()
	d = embedDim
	embedDimMu.Unlock()
	if d <= 0 {
		return 0, fmt.Errorf("p_google_genai: EmbeddingDimension: unknown dimension")
	}
	return d, nil
}

func StatusSummary() string {
	var lines []string
	lines = append(lines, "Backend: "+GoogleGenAIConfig.Backend)
	lines = append(lines, "Text model: "+GoogleGenAIConfig.TextModel)
	lines = append(lines, "Embedding model: "+GoogleGenAIConfig.EmbeddingModel)
	lines = append(lines, "Thinking: "+GoogleGenAIConfig.ThinkingMode)
	lines = append(lines, fmt.Sprintf("Retries: %d extra attempts (429/backoff), base %dms", GoogleGenAIConfig.RetryMax, GoogleGenAIConfig.RetryBaseMillis))
	if GoogleGenAIConfig.Backend == BackendVertex {
		lines = append(lines, "Project: "+GoogleGenAIConfig.Project)
		lines = append(lines, "Location: "+GoogleGenAIConfig.Location)
	} else {
		if GoogleGenAIConfig.APIKey != "" {
			lines = append(lines, "API key: configured in [Plugins.p_google_genai]")
		} else {
			lines = append(lines, "API key: from GOOGLE_API_KEY / GEMINI_API_KEY env")
		}
	}
	return strings.Join(lines, "\n")
}

func genaiClientFor(ctx context.Context) (*genai.Client, error) {
	clientMu.Lock()
	defer clientMu.Unlock()
	if cachedClient != nil {
		return cachedClient, nil
	}
	cc := &genai.ClientConfig{}
	switch GoogleGenAIConfig.Backend {
	case BackendVertex:
		cc.Backend = genai.BackendVertexAI
		cc.Project = GoogleGenAIConfig.Project
		cc.Location = GoogleGenAIConfig.Location
	default:
		cc.Backend = genai.BackendGeminiAPI
		cc.APIKey = GoogleGenAIConfig.APIKey
		if strings.TrimSpace(GoogleGenAIConfig.APIKey) == "" {
			warnedEmptyAPIKey.Do(func() {
				slog.WarnContext(ctx, "p_google_genai: API key not set in loaded config — the config path passed to lago.LoadConfigFromFile must define [Plugins.p_google_genai] with apiKey or api_key (table key must be exactly p_google_genai). If missing, nothing is decoded and the GenAI SDK uses GOOGLE_API_KEY / GEMINI_API_KEY. After changing apiKey, restart the process (client is cached).")
			})
		}
	}
	cli, err := genai.NewClient(ctx, cc)
	if err != nil {
		return nil, fmt.Errorf("p_google_genai: NewClient: %w", err)
	}
	cachedClient = cli
	return cachedClient, nil
}

func baseGenerateConfig(req GenerateRequest) *genai.GenerateContentConfig {
	maxTok := req.MaxOutputTokens
	if maxTok <= 0 {
		maxTok = GoogleGenAIConfig.MaxOutputTokens
	}
	if maxTok <= 0 {
		maxTok = 8192
	}
	temp := float32(GoogleGenAIConfig.TextTemperature)
	if req.Temperature != nil {
		temp = *req.Temperature
	}
	topP := float32(GoogleGenAIConfig.TextTopP)
	cfg := &genai.GenerateContentConfig{
		Temperature:     genai.Ptr(temp),
		TopP:            genai.Ptr(topP),
		TopK:            genai.Ptr(GoogleGenAIConfig.TextTopK),
		MaxOutputTokens: int32(maxTok),
		ThinkingConfig:  genaiThinking(req.Thinking),
	}
	return cfg
}

func systemInstruction(systemPrompt string, th *ThinkingConfig) *genai.Content {
	base := strings.TrimSpace(systemPrompt)
	if base == "" {
		base = applyThinkingDirective("", resolveThinking(th))
	} else {
		base = applyThinkingDirective(base, resolveThinking(th))
	}
	if strings.TrimSpace(base) == "" {
		return nil
	}
	return genai.NewContentFromText(base, "")
}

func genaiThinking(override *ThinkingConfig) *genai.ThinkingConfig {
	th := resolveThinking(override)
	if th.Mode != ThinkingModeEnabled {
		return nil
	}
	b := int32(th.Budget)
	if b <= 0 {
		b = 8192
	}
	return &genai.ThinkingConfig{
		IncludeThoughts: false,
		ThinkingBudget:  genai.Ptr(b),
	}
}

func resolveThinking(override *ThinkingConfig) ThinkingConfig {
	if override == nil {
		return ThinkingConfig{
			Mode:   GoogleGenAIConfig.ThinkingMode,
			Budget: GoogleGenAIConfig.ThinkingBudget,
		}
	}
	mode := strings.ToLower(strings.TrimSpace(override.Mode))
	if mode == "" || mode == ThinkingModeDefault {
		mode = GoogleGenAIConfig.ThinkingMode
	}
	if mode != ThinkingModeEnabled && mode != ThinkingModeDisabled {
		mode = ThinkingModeDisabled
	}
	budget := override.Budget
	if budget <= 0 {
		budget = GoogleGenAIConfig.ThinkingBudget
	}
	return ThinkingConfig{Mode: mode, Budget: budget}
}

func applyThinkingDirective(systemPrompt string, thinking ThinkingConfig) string {
	systemPrompt = strings.TrimSpace(systemPrompt)
	switch thinking.Mode {
	case ThinkingModeEnabled:
		if thinking.Budget > 0 {
			suffix := fmt.Sprintf("\n\nReason carefully before answering. Keep private chain-of-thought out of the user-visible answer. Aim for roughly %d reasoning tokens or less.", thinking.Budget)
			if systemPrompt == "" {
				return strings.TrimSpace(suffix)
			}
			return systemPrompt + suffix
		}
		suffix := "\n\nReason carefully before answering. Keep private chain-of-thought out of the user-visible answer."
		if systemPrompt == "" {
			return strings.TrimSpace(suffix)
		}
		return systemPrompt + suffix
	default:
		suffix := "\n\nDo not emit chain-of-thought or hidden-reasoning tags. Return only the final answer."
		if systemPrompt == "" {
			return strings.TrimSpace(suffix)
		}
		return systemPrompt + suffix
	}
}

func runGenerateStream(ctx context.Context, cli *genai.Client, model string, contents []*genai.Content, cfg *genai.GenerateContentConfig, onToken func(string) error) (string, error) {
	if onToken == nil {
		logGenerateRequest(ctx, "generate_content", model, cfg, contents, false)
		resp, err := withGenAIRetryResp(ctx, "generate_content", func() (*genai.GenerateContentResponse, error) {
			return cli.Models.GenerateContent(ctx, model, contents, cfg)
		})
		if err != nil {
			slog.WarnContext(ctx, "p_google_genai: GenerateContent error", "model", model, "error", err.Error())
			return "", err
		}
		out := strings.TrimSpace(resp.Text())
		if out == "" {
			return "", fmt.Errorf("p_google_genai: GenerateTextStream: empty response")
		}
		return out, nil
	}
	logGenerateRequest(ctx, "generate_content_stream", model, cfg, contents, true)
	retries := effectiveRetryCount()
	attempts := 1 + retries
	var lastStreamErr error
	for i := 0; i < attempts; i++ {
		if i > 0 {
			delay := retryDelay(i - 1)
			slog.InfoContext(ctx, "p_google_genai: retrying stream (no tokens delivered yet)",
				"op", "generate_content_stream",
				"attempt", i+1,
				"maxAttempts", attempts,
				"delay", delay.String(),
				"error", lastStreamErr.Error())
			select {
			case <-ctx.Done():
				return "", fmt.Errorf("%w: last genai error: %v", ctx.Err(), lastStreamErr)
			case <-time.After(delay):
			}
		}
		out, emitted, err := consumeGenerateContentStream(ctx, cli, model, contents, cfg, onToken)
		if err == nil {
			if out == "" {
				return "", fmt.Errorf("p_google_genai: GenerateTextStream: empty response")
			}
			return out, nil
		}
		lastStreamErr = err
		if emitted || !isRetryableGenAIError(err) {
			slog.WarnContext(ctx, "p_google_genai: GenerateContentStream error", "model", model, "error", err.Error())
			return "", err
		}
	}
	slog.WarnContext(ctx, "p_google_genai: GenerateContentStream error (exhausted retries)", "model", model, "error", lastStreamErr.Error())
	return "", lastStreamErr
}

func consumeGenerateContentStream(ctx context.Context, cli *genai.Client, model string, contents []*genai.Content, cfg *genai.GenerateContentConfig, onToken func(string) error) (out string, emittedToken bool, err error) {
	var full strings.Builder
	for resp, err := range cli.Models.GenerateContentStream(ctx, model, contents, cfg) {
		if err != nil {
			return "", emittedToken, err
		}
		if resp == nil {
			continue
		}
		piece := resp.Text()
		if piece == "" {
			continue
		}
		full.WriteString(piece)
		if onToken != nil {
			emittedToken = true
			if err := onToken(piece); err != nil {
				return "", emittedToken, err
			}
		}
	}
	return strings.TrimSpace(full.String()), emittedToken, nil
}

func embeddingTaskPrefix(task EmbedTask) string {
	switch task {
	case EmbedTaskSearchQuery:
		return "search_query: "
	case EmbedTaskSearchDocument, EmbedTaskVisionDocument:
		return "search_document: "
	default:
		return ""
	}
}

func embedTaskType(task EmbedTask) string {
	switch task {
	case EmbedTaskSearchQuery:
		return "RETRIEVAL_QUERY"
	default:
		return "RETRIEVAL_DOCUMENT"
	}
}

func decodeJSONBlob(raw string, target any) error {
	blob := extractJSONBlob(raw)
	if blob == "" {
		return fmt.Errorf("empty json blob")
	}
	return json.Unmarshal([]byte(blob), target)
}

func extractJSONBlob(raw string) string {
	raw = strings.TrimSpace(raw)
	obj := strings.Index(raw, "{")
	arr := strings.Index(raw, "[")
	switch {
	case obj >= 0 && (arr < 0 || obj < arr):
		end := strings.LastIndex(raw, "}")
		if end > obj {
			return strings.TrimSpace(raw[obj : end+1])
		}
	case arr >= 0:
		end := strings.LastIndex(raw, "]")
		if end > arr {
			return strings.TrimSpace(raw[arr : end+1])
		}
	}
	return raw
}

func normalizeVector(values []float32) []float32 {
	var sum float64
	for _, v := range values {
		sum += float64(v * v)
	}
	if sum == 0 {
		out := make([]float32, len(values))
		copy(out, values)
		return out
	}
	norm := float32(1.0 / math.Sqrt(sum))
	out := make([]float32, len(values))
	for i, v := range values {
		out[i] = v * norm
	}
	return out
}
