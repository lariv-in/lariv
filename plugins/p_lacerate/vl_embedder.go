package p_lacerate

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"

	"github.com/lariv-in/lago/plugins/p_filesystem"
	"github.com/pgvector/pgvector-go"
	"gorm.io/gorm"
)

// IntelEmbeddingDim is the embedding width stored in [Intel.Embedding] and [TargetOfInterest.Embedding].
// It must match the configured [VLEmbedder] output; the GORM column type uses the same numeric literal.
const IntelEmbeddingDim = 1024

// VLEmbedder calls an external multimodal embedding service (e.g. [GenAIVLEmbedder] via Gemini API).
// Implementations must return a slice of length [IntelEmbeddingDim] on success.
type VLEmbedder interface {
	Embed(ctx context.Context, text string, images ...[]byte) ([]float32, error)
}

var (
	vlEmbedderMu      sync.RWMutex
	defaultVLEmbedder VLEmbedder // nil until [RegisterVLEmbedder]
	loggedNilEmbedder sync.Once  // one-time hint when apiKey is unset
)

// RegisterVLEmbedder sets the package-default embedder used after Reddit (and other) ingest creates an [Intel].
// Pass nil to clear.
func RegisterVLEmbedder(e VLEmbedder) {
	vlEmbedderMu.Lock()
	defaultVLEmbedder = e
	vlEmbedderMu.Unlock()
}

func vlEmbedder() VLEmbedder {
	vlEmbedderMu.RLock()
	defer vlEmbedderMu.RUnlock()
	return defaultVLEmbedder
}

// prepareIntelEmbeddingForSave sets [Intel.Embedding] on the struct before INSERT/UPDATE.
// Without a [VLEmbedder], stores a zero vector so NOT NULL is satisfied. On embed failure or wrong dimension, returns an error.
func prepareIntelEmbeddingForSave(ctx context.Context, db *gorm.DB, intel *Intel) error {
	if intel == nil {
		return nil
	}
	e := vlEmbedder()
	if e == nil {
		loggedNilEmbedder.Do(func() {
			slog.Info("lacerate: VLEmbedder not configured (set [p_lacerate.geminiEmbedding] apiKey in config); Intel embeddings use zero vector")
		})
		intel.Embedding = pgvector.NewVector(make([]float32, IntelEmbeddingDim))
		return nil
	}
	var images [][]byte
	if intel.PreviewImageID != nil {
		var node p_filesystem.VNode
		if err := db.First(&node, *intel.PreviewImageID).Error; err != nil {
			slog.Error("lacerate: load intel preview vnode for embed", "error", err, "intel_id", intel.ID, "preview_image_id", *intel.PreviewImageID)
		} else if b, err := vnodeFileBytes(&node); err != nil {
			slog.Error("lacerate: read intel preview file for embed", "error", err, "intel_id", intel.ID, "vnode_id", node.ID)
		} else if len(b) > 0 {
			images = append(images, b)
		}
	}
	vec, err := e.Embed(ctx, intel.Content, images...)
	if err != nil {
		slog.Error("lacerate: vl embed intel", "error", err, "intel_id", intel.ID)
		return fmt.Errorf("lacerate: vl embed intel: %w", err)
	}
	if len(vec) != IntelEmbeddingDim {
		slog.Error("lacerate: vl embed intel wrong dimension", "got", len(vec), "want", IntelEmbeddingDim, "intel_id", intel.ID)
		return fmt.Errorf("lacerate: vl embed intel: got dimension %d, want %d", len(vec), IntelEmbeddingDim)
	}
	intel.Embedding = pgvector.NewVector(vec)
	slog.Info("lacerate: vl embed intel success", "intel_id", intel.ID, "dim", len(vec))
	return nil
}

// prepareTargetOfInterestEmbeddingForSave sets [TargetOfInterest.Embedding] on the struct before INSERT/UPDATE.
// With no [VLEmbedder] or empty [TargetOfInterest.String], leaves embedding cleared or unchanged. On embed failure or wrong dimension, returns an error without updating the embedding.
func prepareTargetOfInterestEmbeddingForSave(ctx context.Context, a *TargetOfInterest) error {
	if a == nil {
		return nil
	}
	e := vlEmbedder()
	if e == nil {
		return nil
	}
	text := a.String()
	if text == "" {
		a.Embedding = nil
		return nil
	}
	vec, err := e.Embed(ctx, text)
	if err != nil {
		slog.Error("lacerate: vl embed Target of Interest", "error", err, "target_of_interest_id", a.ID)
		return fmt.Errorf("lacerate: vl embed Target of Interest: %w", err)
	}
	if len(vec) != IntelEmbeddingDim {
		slog.Error("lacerate: vl embed Target of Interest wrong dimension", "got", len(vec), "want", IntelEmbeddingDim, "target_of_interest_id", a.ID)
		return fmt.Errorf("lacerate: vl embed Target of Interest: got dimension %d, want %d", len(vec), IntelEmbeddingDim)
	}
	v := pgvector.NewVector(vec)
	a.Embedding = &v
	slog.Info("lacerate: vl embed Target of Interest success", "target_of_interest_id", a.ID, "dim", len(vec))
	return nil
}

// prepareLookupEmbeddingForSave sets [Lookup.Embedding] from [Lookup.Content] before INSERT/UPDATE.
// Empty content stores a zero vector. When no [VLEmbedder] is registered and content is non-empty, the existing embedding is left unchanged.
// On embed failure or wrong dimension, returns an error and does not overwrite [Lookup.Embedding].
func prepareLookupEmbeddingForSave(ctx context.Context, l *Lookup) error {
	if l == nil {
		return nil
	}
	text := strings.TrimSpace(l.Content)
	if text == "" {
		l.Embedding = pgvector.NewVector(make([]float32, IntelEmbeddingDim))
		return nil
	}
	e := vlEmbedder()
	if e == nil {
		return nil
	}
	vec, err := e.Embed(ctx, text)
	if err != nil {
		slog.Error("lacerate: vl embed lookup", "error", err, "lookup_id", l.ID)
		return fmt.Errorf("lacerate: vl embed lookup: %w", err)
	}
	if len(vec) != IntelEmbeddingDim {
		slog.Error("lacerate: vl embed lookup wrong dimension", "got", len(vec), "want", IntelEmbeddingDim, "lookup_id", l.ID)
		return fmt.Errorf("lacerate: vl embed lookup: got dimension %d, want %d", len(vec), IntelEmbeddingDim)
	}
	l.Embedding = pgvector.NewVector(vec)
	slog.Info("lacerate: vl embed lookup success", "lookup_id", l.ID, "dim", len(vec))
	return nil
}
