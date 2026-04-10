package p_lacerate

import (
	"context"
	"os"
	"strings"
	"testing"
)

const lacerateTestGeminiAPIKeyEnv = "LACERATE_TEST_GEMINI_API_KEY"

func TestGenAIVLEmbedderReturnsNonZeroEmbedding(t *testing.T) {
	apiKey := strings.TrimSpace(os.Getenv(lacerateTestGeminiAPIKeyEnv))
	if apiKey == "" {
		t.Skipf("%s is not set", lacerateTestGeminiAPIKeyEnv)
	}

	emb, err := NewGenAIVLEmbedder(context.Background(), apiKey, defaultGeminiEmbeddingModel)
	if err != nil {
		t.Fatalf("NewGenAIVLEmbedder failed: %v", err)
	}

	vec, err := emb.Embed(context.Background(), "task: search result | query: latest world news")
	if err != nil {
		t.Fatalf("Embed failed: %v", err)
	}
	t.Logf("embedding=%v", vec)

	if len(vec) != IntelEmbeddingDim {
		t.Fatalf("expected dim %d, got %d", IntelEmbeddingDim, len(vec))
	}
	if nonZero, absSum := embeddingStats(vec); nonZero == 0 || absSum == 0 {
		t.Fatalf("expected non-zero embedding, got non_zero=%d abs_sum=%v", nonZero, absSum)
	}
}

func TestIntelCreateEndToEndWithGenAIAndPostgres(t *testing.T) {
	apiKey := strings.TrimSpace(os.Getenv(lacerateTestGeminiAPIKeyEnv))
	if apiKey == "" {
		t.Skipf("%s is not set", lacerateTestGeminiAPIKeyEnv)
	}

	db := lacerateTestPostgresDB(t)

	emb, err := NewGenAIVLEmbedder(context.Background(), apiKey, defaultGeminiEmbeddingModel)
	if err != nil {
		t.Fatalf("NewGenAIVLEmbedder failed: %v", err)
	}

	previous := vlEmbedder()
	RegisterVLEmbedder(emb)
	t.Cleanup(func() {
		RegisterVLEmbedder(previous)
	})

	src := Source{Name: "e2e", Kind: "reddit"}
	if err := db.Create(&src).Error; err != nil {
		t.Fatalf("Create source failed: %v", err)
	}

	dedup := "e2e-dedup"
	intel := Intel{
		SourceID:  src.ID,
		DedupHash: &dedup,
		Content:   "task: search result | query: latest world news about geopolitics",
	}
	if err := db.Create(&intel).Error; err != nil {
		t.Fatalf("Create intel failed: %v", err)
	}

	var got Intel
	if err := db.First(&got, intel.ID).Error; err != nil {
		t.Fatalf("Load intel failed: %v", err)
	}

	vec := got.Embedding.Slice()
	if len(vec) != IntelEmbeddingDim {
		t.Fatalf("expected embedding dim %d, got %d", IntelEmbeddingDim, len(vec))
	}
	nonZero, absSum := embeddingStats(vec)
	t.Logf("persisted embedding=%v", vec)
	t.Logf("persisted embedding stats: non_zero=%d abs_sum=%v", nonZero, absSum)
	if nonZero == 0 || absSum == 0 {
		t.Fatalf("expected persisted non-zero embedding, got non_zero=%d abs_sum=%v", nonZero, absSum)
	}
}
