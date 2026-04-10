package p_lacerate

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"net/http"
	"strings"

	"google.golang.org/genai"
)

const defaultGeminiEmbeddingModel = "gemini-embedding-2-preview"

// geminiEmbeddingModelMaxImages is the limit documented for gemini-embedding-2-preview image parts per request.
const geminiEmbeddingModelMaxImages = 6

// GenAIVLEmbedder implements [VLEmbedder] using the Gemini API via [google.golang.org/genai]
// (same client stack as Google ADK Go quickstart). It uses gemini-embedding-2-preview by default
// for multimodal (text + image) inputs with [IntelEmbeddingDim] output.
type GenAIVLEmbedder struct {
	client *genai.Client
	model  string
}

func allZeroEmbedding(vec []float32) bool {
	for _, v := range vec {
		if math.Abs(float64(v)) > 1e-12 {
			return false
		}
	}
	return len(vec) > 0
}

// NewGenAIVLEmbedder returns a [VLEmbedder] backed by the Gemini Developer API. model may be empty
// to use [defaultGeminiEmbeddingModel].
func NewGenAIVLEmbedder(ctx context.Context, apiKey, model string) (*GenAIVLEmbedder, error) {
	if strings.TrimSpace(apiKey) == "" {
		err := fmt.Errorf("p_lacerate: gemini embedding api key is empty")
		slog.Error("lacerate: new genai VL embedder", "error", err)
		return nil, err
	}
	if strings.TrimSpace(model) == "" {
		model = defaultGeminiEmbeddingModel
	}
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		slog.Error("lacerate: genai new client for embedder", "error", err)
		return nil, err
	}
	return &GenAIVLEmbedder{client: client, model: model}, nil
}

// Embed implements [VLEmbedder].
func (e *GenAIVLEmbedder) Embed(ctx context.Context, text string, images ...[]byte) ([]float32, error) {
	if e == nil || e.client == nil {
		err := fmt.Errorf("p_lacerate: nil GenAIVLEmbedder")
		slog.Error("lacerate: genai embed", "error", err)
		return nil, err
	}
	var parts []*genai.Part
	if t := strings.TrimSpace(text); t != "" {
		parts = append(parts, genai.NewPartFromText(t))
	}
	added := 0
	for _, img := range images {
		if added >= geminiEmbeddingModelMaxImages {
			break
		}
		if len(img) == 0 {
			continue
		}
		mime := http.DetectContentType(img)
		switch mime {
		case "image/png", "image/jpeg", "image/webp", "image/gif":
			// Gemini multimodal embedding accepts these; Reddit/Twitter previews are often WebP.
		default:
			continue
		}
		parts = append(parts, genai.NewPartFromBytes(img, mime))
		added++
	}
	if len(parts) == 0 {
		err := fmt.Errorf("p_lacerate: genai embed needs non-empty text and/or at least one PNG/JPEG image")
		slog.Error("lacerate: genai embed", "error", err)
		return nil, err
	}
	contents := []*genai.Content{genai.NewContentFromParts(parts, genai.RoleUser)}
	dim := int32(IntelEmbeddingDim)
	cfg := &genai.EmbedContentConfig{OutputDimensionality: &dim}
	res, err := e.client.Models.EmbedContent(ctx, e.model, contents, cfg)
	if err != nil {
		slog.Error("lacerate: genai embed content", "error", err)
		return nil, err
	}
	if len(res.Embeddings) == 0 || res.Embeddings[0] == nil {
		err := fmt.Errorf("p_lacerate: genai embed returned no embeddings")
		slog.Error("lacerate: genai embed", "error", err)
		return nil, err
	}
	values := res.Embeddings[0].Values
	if len(values) != IntelEmbeddingDim {
		err := fmt.Errorf("p_lacerate: genai embed got dimension %d, want %d", len(values), IntelEmbeddingDim)
		slog.Error("lacerate: genai embed", "error", err, "got", len(values), "want", IntelEmbeddingDim)
		return nil, err
	}
	if allZeroEmbedding(values) {
		err := fmt.Errorf("p_lacerate: genai embed returned an all-zero vector")
		slog.Error("lacerate: genai embed", "error", err, "model", e.model, "dim", len(values))
		return nil, err
	}
	return values, nil
}
