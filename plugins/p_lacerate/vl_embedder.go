package p_lacerate

import (
	"context"
	"fmt"
	"log/slog"
	"math"
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

// intelZeroEmbedding is a valid NOT NULL placeholder when VL embedding fails or is unavailable.
func intelZeroEmbedding() pgvector.Vector {
	return pgvector.NewVector(make([]float32, IntelEmbeddingDim))
}

func embeddingStats(vec []float32) (nonZero int, absSum float64) {
	for _, v := range vec {
		f := math.Abs(float64(v))
		absSum += f
		if f > 1e-12 {
			nonZero++
		}
	}
	return nonZero, absSum
}

// prepareIntelEmbeddingForSave sets [Intel.Embedding] on the struct before INSERT/UPDATE.
// Without a [VLEmbedder], on embed failure, or on wrong dimension, uses [intelZeroEmbedding] so NOT NULL is satisfied.
func prepareIntelEmbeddingForSave(ctx context.Context, db *gorm.DB, intel *Intel) {
	if intel == nil {
		return
	}
	e := vlEmbedder()
	if e == nil {
		loggedNilEmbedder.Do(func() {
			slog.Info("lacerate: VLEmbedder not configured (set [p_lacerate.geminiEmbedding] apiKey in config); Intel embeddings use zero vector")
		})
		intel.Embedding = intelZeroEmbedding()
		return
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
		intel.Embedding = intelZeroEmbedding()
		return
	}
	if len(vec) != IntelEmbeddingDim {
		slog.Error("lacerate: vl embed intel wrong dimension", "got", len(vec), "want", IntelEmbeddingDim, "intel_id", intel.ID)
		intel.Embedding = intelZeroEmbedding()
		return
	}
	intel.Embedding = pgvector.NewVector(vec)
	nonZero, absSum := embeddingStats(vec)
	slog.Info("lacerate: vl embed intel success", "intel_id", intel.ID, "dim", len(vec), "non_zero", nonZero, "abs_sum", absSum)
}

func targetOfInterestEmbeddingText(a *TargetOfInterest) string {
	if a == nil {
		return ""
	}
	var b strings.Builder
	if t := strings.TrimSpace(a.Name); t != "" {
		fmt.Fprintf(&b, "# %s\n\n", t)
	}
	if t := strings.TrimSpace(a.Type); t != "" {
		fmt.Fprintf(&b, "**Type:** %s\n\n", t)
	}
	if t := strings.TrimSpace(a.Description); t != "" {
		b.WriteString(t)
		b.WriteString("\n\n")
	}
	if t := strings.TrimSpace(a.Content); t != "" {
		b.WriteString(t)
	}
	return strings.TrimSpace(b.String())
}

// prepareTargetOfInterestEmbeddingForSave sets [TargetOfInterest.Embedding] on the struct before INSERT/UPDATE.
func prepareTargetOfInterestEmbeddingForSave(ctx context.Context, a *TargetOfInterest) {
	e := vlEmbedder()
	if e == nil || a == nil {
		return
	}
	text := targetOfInterestEmbeddingText(a)
	if text == "" {
		a.Embedding = nil
		return
	}
	vec, err := e.Embed(ctx, text)
	if err != nil {
		slog.Error("lacerate: vl embed Target of Interest", "error", err, "target_of_interest_id", a.ID)
		return
	}
	if len(vec) != IntelEmbeddingDim {
		slog.Error("lacerate: vl embed Target of Interest wrong dimension", "got", len(vec), "want", IntelEmbeddingDim, "target_of_interest_id", a.ID)
		return
	}
	v := pgvector.NewVector(vec)
	a.Embedding = &v
	nonZero, absSum := embeddingStats(vec)
	slog.Info("lacerate: vl embed Target of Interest success", "target_of_interest_id", a.ID, "dim", len(vec), "non_zero", nonZero, "abs_sum", absSum)
}

// prepareLookupEmbeddingForSave sets [Lookup.Embedding] from [Lookup.Content] before INSERT/UPDATE.
// When no [VLEmbedder] is registered and content is non-empty, the existing embedding value is left unchanged
// (for updates this keeps the loaded column; [Lookup.BeforeCreate] still supplies a zero vector on create).
func prepareLookupEmbeddingForSave(ctx context.Context, l *Lookup) {
	if l == nil {
		return
	}
	text := strings.TrimSpace(l.Content)
	if text == "" {
		l.Embedding = intelZeroEmbedding()
		return
	}
	e := vlEmbedder()
	if e == nil {
		return
	}
	vec, err := e.Embed(ctx, text)
	if err != nil {
		slog.Error("lacerate: vl embed lookup", "error", err, "lookup_id", l.ID)
		l.Embedding = intelZeroEmbedding()
		return
	}
	if len(vec) != IntelEmbeddingDim {
		slog.Error("lacerate: vl embed lookup wrong dimension", "got", len(vec), "want", IntelEmbeddingDim, "lookup_id", l.ID)
		l.Embedding = intelZeroEmbedding()
		return
	}
	l.Embedding = pgvector.NewVector(vec)
	nonZero, absSum := embeddingStats(vec)
	slog.Info("lacerate: vl embed lookup success", "lookup_id", l.ID, "dim", len(vec), "non_zero", nonZero, "abs_sum", absSum)
}
