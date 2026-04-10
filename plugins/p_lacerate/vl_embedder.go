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

func gormSkipHooks(tx *gorm.DB) *gorm.DB {
	// Use a fresh statement for hook-triggered embedding writes so an AfterSave update
	// does not inherit the original Create/Save statement state.
	return tx.Session(&gorm.Session{NewDB: true, SkipHooks: true})
}

func intelEmbeddingPersist(db *gorm.DB, intelID uint, v pgvector.Vector) error {
	if intelID == 0 {
		err := fmt.Errorf("intel id is zero")
		slog.Error("lacerate: intel embedding persist", "error", err)
		return err
	}
	if db.Dialector.Name() == "postgres" {
		lit := formatFloat32VectorForPG(v.Slice())
		return gormSkipHooks(db).Exec(`UPDATE intels SET embedding = ?::vector WHERE id = ?`, lit, intelID).Error
	}
	return gormSkipHooks(db).Model(&Intel{}).Where("id = ?", intelID).Update("embedding", v).Error
}

func targetOfInterestEmbeddingPersist(db *gorm.DB, rowID uint, v *pgvector.Vector) error {
	if rowID == 0 {
		err := fmt.Errorf("Target of Interest id is zero")
		slog.Error("lacerate: Target of Interest embedding persist", "error", err)
		return err
	}
	if db.Dialector.Name() == "postgres" {
		if v == nil {
			return gormSkipHooks(db).Exec(`UPDATE targets_of_interest SET embedding = NULL WHERE id = ?`, rowID).Error
		}
		lit := formatFloat32VectorForPG(v.Slice())
		return gormSkipHooks(db).Exec(`UPDATE targets_of_interest SET embedding = ?::vector WHERE id = ?`, lit, rowID).Error
	}
	return gormSkipHooks(db).Model(&TargetOfInterest{}).Where("id = ?", rowID).Update("embedding", v).Error
}

func lookupEmbeddingPersist(db *gorm.DB, lookupID uint, v pgvector.Vector) error {
	if lookupID == 0 {
		err := fmt.Errorf("lookup id is zero")
		slog.Error("lacerate: lookup embedding persist", "error", err)
		return err
	}
	if db.Dialector.Name() == "postgres" {
		lit := formatFloat32VectorForPG(v.Slice())
		return gormSkipHooks(db).Exec(`UPDATE lookups SET embedding = ?::vector WHERE id = ?`, lit, lookupID).Error
	}
	return gormSkipHooks(db).Model(&Lookup{}).Where("id = ?", lookupID).Update("embedding", v).Error
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

// applyIntelEmbedding runs the registered [VLEmbedder] and persists [Intel.Embedding].
// On failure or missing embedder it logs and stores [intelZeroEmbedding] so NOT NULL is satisfied.
func applyIntelEmbedding(ctx context.Context, db *gorm.DB, intel *Intel) {
	if intel == nil {
		return
	}
	e := vlEmbedder()
	if e == nil {
		loggedNilEmbedder.Do(func() {
			slog.Info("lacerate: VLEmbedder not configured (set [p_lacerate.geminiEmbedding] apiKey in config); Intel embeddings are not computed and stay zero")
		})
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
		if err2 := intelEmbeddingPersist(db, intel.ID, intel.Embedding); err2 != nil {
			slog.Error("lacerate: save intel zero embedding after embed error", "error", err2, "intel_id", intel.ID)
		}
		return
	}
	if len(vec) != IntelEmbeddingDim {
		slog.Error("lacerate: vl embed intel wrong dimension", "got", len(vec), "want", IntelEmbeddingDim, "intel_id", intel.ID)
		intel.Embedding = intelZeroEmbedding()
		if err2 := intelEmbeddingPersist(db, intel.ID, intel.Embedding); err2 != nil {
			slog.Error("lacerate: save intel zero embedding after dim error", "error", err2, "intel_id", intel.ID)
		}
		return
	}
	intel.Embedding = pgvector.NewVector(vec)
	nonZero, absSum := embeddingStats(vec)
	slog.Info("lacerate: vl embed intel success", "intel_id", intel.ID, "dim", len(vec), "non_zero", nonZero, "abs_sum", absSum)
	if err := intelEmbeddingPersist(db, intel.ID, intel.Embedding); err != nil {
		slog.Error("lacerate: save intel embedding", "error", err, "intel_id", intel.ID)
		return
	}
	slog.Info("lacerate: save intel embedding ok", "intel_id", intel.ID)
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

// applyTargetOfInterestEmbedding runs the registered [VLEmbedder] and persists [TargetOfInterest.Embedding].
func applyTargetOfInterestEmbedding(ctx context.Context, db *gorm.DB, a *TargetOfInterest) {
	e := vlEmbedder()
	if e == nil || a == nil {
		return
	}
	text := targetOfInterestEmbeddingText(a)
	if text == "" {
		a.Embedding = nil
		if err := targetOfInterestEmbeddingPersist(db, a.ID, nil); err != nil {
			slog.Error("lacerate: clear Target of Interest embedding", "error", err, "target_of_interest_id", a.ID)
		}
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
	if err := targetOfInterestEmbeddingPersist(db, a.ID, a.Embedding); err != nil {
		slog.Error("lacerate: save Target of Interest embedding", "error", err, "target_of_interest_id", a.ID)
		a.Embedding = nil
		return
	}
	slog.Info("lacerate: save Target of Interest embedding ok", "target_of_interest_id", a.ID)
}

// applyLookupEmbedding runs the registered [VLEmbedder] and persists [Lookup.Embedding] (text-only).
func applyLookupEmbedding(ctx context.Context, db *gorm.DB, l *Lookup) {
	if l == nil {
		return
	}
	text := strings.TrimSpace(l.Content)
	if text == "" {
		l.Embedding = intelZeroEmbedding()
		if err := lookupEmbeddingPersist(db, l.ID, l.Embedding); err != nil {
			slog.Error("lacerate: save lookup zero embedding (empty content)", "error", err, "lookup_id", l.ID)
		}
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
		if err2 := lookupEmbeddingPersist(db, l.ID, l.Embedding); err2 != nil {
			slog.Error("lacerate: save lookup zero embedding after embed error", "error", err2, "lookup_id", l.ID)
		}
		return
	}
	if len(vec) != IntelEmbeddingDim {
		slog.Error("lacerate: vl embed lookup wrong dimension", "got", len(vec), "want", IntelEmbeddingDim, "lookup_id", l.ID)
		l.Embedding = intelZeroEmbedding()
		if err2 := lookupEmbeddingPersist(db, l.ID, l.Embedding); err2 != nil {
			slog.Error("lacerate: save lookup zero embedding after dim error", "error", err2, "lookup_id", l.ID)
		}
		return
	}
	l.Embedding = pgvector.NewVector(vec)
	nonZero, absSum := embeddingStats(vec)
	slog.Info("lacerate: vl embed lookup success", "lookup_id", l.ID, "dim", len(vec), "non_zero", nonZero, "abs_sum", absSum)
	if err := lookupEmbeddingPersist(db, l.ID, l.Embedding); err != nil {
		slog.Error("lacerate: save lookup embedding", "error", err, "lookup_id", l.ID)
		return
	}
	slog.Info("lacerate: save lookup embedding ok", "lookup_id", l.ID)
}
