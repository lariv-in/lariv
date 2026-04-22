package p_google_genai

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"google.golang.org/genai"
)

const maxLogBytesGenerate = 256 * 1024
const maxLogBytesEmbed = 128 * 1024

type loggedContent struct {
	Role string `json:"role"`
	Text string `json:"text"`
}

type loggedGeneratePayload struct {
	Op                string          `json:"op"`
	Backend           string          `json:"backend"`
	Model             string          `json:"model"`
	Stream            bool            `json:"stream"`
	Temperature       *float32        `json:"temperature,omitempty"`
	TopP              *float32        `json:"topP,omitempty"`
	TopK              *float32        `json:"topK,omitempty"`
	MaxOutputTokens   int32           `json:"maxOutputTokens,omitempty"`
	ThinkingConfig    any             `json:"thinkingConfig,omitempty"`
	SystemInstruction string          `json:"systemInstruction,omitempty"`
	Messages          []loggedContent `json:"messages"`
}

type loggedEmbedPayload struct {
	Op          string `json:"op"`
	Backend     string `json:"backend"`
	Model       string `json:"model"`
	EmbedTask   string `json:"embedTask"`
	APITaskType string `json:"apiTaskType"`
	InputBytes  int    `json:"inputBytes"`
	Input       string `json:"input"`
}

func logGenerateRequest(ctx context.Context, op string, model string, cfg *genai.GenerateContentConfig, contents []*genai.Content, stream bool) {
	payload := loggedGeneratePayload{
		Op:       op,
		Backend:  GoogleGenAIConfig.Backend,
		Model:    model,
		Stream:   stream,
		Messages: contentsToLogged(contents),
	}
	if cfg != nil {
		payload.Temperature = cfg.Temperature
		payload.TopP = cfg.TopP
		payload.TopK = cfg.TopK
		payload.MaxOutputTokens = cfg.MaxOutputTokens
		payload.ThinkingConfig = cfg.ThinkingConfig
		if cfg.SystemInstruction != nil {
			payload.SystemInstruction = truncateForLog(joinContentText(cfg.SystemInstruction), maxLogBytesGenerate)
		}
	}
	for i := range payload.Messages {
		payload.Messages[i].Text = truncateForLog(payload.Messages[i].Text, maxLogBytesGenerate)
	}
	b, err := json.Marshal(payload)
	if err != nil {
		slog.WarnContext(ctx, "p_google_genai: log marshal failed", "error", err.Error())
		return
	}
	slog.InfoContext(ctx, "p_google_genai: outgoing request", "payload", string(b))
}

func logEmbedRequest(ctx context.Context, task EmbedTask, input string, apiTaskType string) {
	payload := loggedEmbedPayload{
		Op:          "embed_content",
		Backend:     GoogleGenAIConfig.Backend,
		Model:       GoogleGenAIConfig.EmbeddingModel,
		EmbedTask:   string(task),
		APITaskType: apiTaskType,
		InputBytes:  len(input),
		Input:       truncateForLog(input, maxLogBytesEmbed),
	}
	b, err := json.Marshal(payload)
	if err != nil {
		slog.WarnContext(ctx, "p_google_genai: log marshal failed", "error", err.Error())
		return
	}
	slog.InfoContext(ctx, "p_google_genai: outgoing request", "payload", string(b))
}

func contentsToLogged(contents []*genai.Content) []loggedContent {
	if len(contents) == 0 {
		return nil
	}
	out := make([]loggedContent, 0, len(contents))
	for _, c := range contents {
		if c == nil {
			continue
		}
		out = append(out, loggedContent{
			Role: c.Role,
			Text: joinContentText(c),
		})
	}
	return out
}

func joinContentText(c *genai.Content) string {
	if c == nil {
		return ""
	}
	var b strings.Builder
	for _, p := range c.Parts {
		if p == nil {
			continue
		}
		b.WriteString(p.Text)
	}
	return b.String()
}

func truncateForLog(s string, maxBytes int) string {
	if maxBytes <= 0 || len(s) <= maxBytes {
		return s
	}
	return s[:maxBytes] + fmt.Sprintf("…[truncated total=%d bytes]", len(s))
}
