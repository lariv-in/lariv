package p_lacerate

import (
	"context"

	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_filesystem"
	"github.com/pgvector/pgvector-go"
	"gorm.io/gorm"
)

type Intel struct {
	gorm.Model
	// uniqueIndex name is shared with DedupHash so GORM creates one composite UNIQUE(source_id, dedup_hash), not a unique source_id alone.
	SourceID *uint   `gorm:"uniqueIndex:idx_intel_source_dedup"`
	Source   *Source `gorm:"foreignKey:SourceID;constraint:OnDelete:SET NULL"`
	// DedupHash is a fixed-width digest for ingest deduplication (e.g. Reddit: SHA-256 hex of post.ID); nil for rows with no dedupe key.
	DedupHash *string `gorm:"size:64;uniqueIndex:idx_intel_source_dedup"`
	// Content is canonical markdown for display and VL text input.
	Content string `gorm:"type:text;not null;default:''"`
	// PreviewImageID optionally points at a copied thumbnail in [p_filesystem] (see [persistRedditPreviewImage]).
	PreviewImageID *uint
	PreviewImage   p_filesystem.VNode `gorm:"foreignKey:PreviewImageID"`
	// Embedding is pgvector storage; dimension must match [IntelEmbeddingDim]. Set in [Intel.BeforeSave] via [prepareIntelEmbeddingForSave].
	Embedding pgvector.Vector `gorm:"type:vector(1024);not null"`
}

// BeforeSave fills [Intel.Embedding] before INSERT/UPDATE ([prepareIntelEmbeddingForSave]).
func (i *Intel) BeforeSave(tx *gorm.DB) error {
	if tx.Statement.SkipHooks {
		return nil
	}
	return prepareIntelEmbeddingForSave(context.Background(), tx, i)
}

func init() {
	lago.OnDBInit("p_lacerate.intel_model", func(db *gorm.DB) *gorm.DB {
		lago.RegisterModel[Intel](db)
		return db
	})
}
