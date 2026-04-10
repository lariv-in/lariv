package p_lacerate

import (
	"context"
	"log/slog"

	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_filesystem"
	"github.com/pgvector/pgvector-go"
	"gorm.io/gorm"
)

// backfillIntelNullContentBeforeMigrate sets legacy NULL content to ” so AutoMigrate can add NOT NULL.
func backfillIntelNullContentBeforeMigrate(db *gorm.DB) {
	if !db.Migrator().HasTable("intels") {
		return
	}
	if !db.Migrator().HasColumn(&Intel{}, "Content") {
		return
	}
	if err := db.Exec(`UPDATE intels SET content = '' WHERE content IS NULL`).Error; err != nil {
		slog.Error("lacerate: backfill intel null content before migrate", "error", err)
	}
}

// backfillIntelNullEmbeddingBeforeMigrate sets NULL embeddings to a zero vector so AutoMigrate can add NOT NULL.
func backfillIntelNullEmbeddingBeforeMigrate(db *gorm.DB) {
	if !db.Migrator().HasTable("intels") {
		return
	}
	if !db.Migrator().HasColumn(&Intel{}, "Embedding") {
		return
	}
	if db.Dialector.Name() != "postgres" {
		return
	}
	lit := intelZeroVectorTextForPG()
	if err := db.Exec(`UPDATE intels SET embedding = ?::vector WHERE embedding IS NULL`, lit).Error; err != nil {
		slog.Error("lacerate: backfill intel null embedding before migrate", "error", err)
	}
}

func intelZeroVectorTextForPG() string {
	b := make([]byte, 0, 2+IntelEmbeddingDim*2)
	b = append(b, '[')
	for i := range IntelEmbeddingDim {
		if i > 0 {
			b = append(b, ',')
		}
		b = append(b, '0')
	}
	b = append(b, ']')
	return string(b)
}

type Intel struct {
	gorm.Model
	// uniqueIndex name is shared with DedupHash so GORM creates one composite UNIQUE(source_id, dedup_hash), not a unique source_id alone.
	SourceID uint   `gorm:"not null;uniqueIndex:idx_intel_source_dedup"`
	Source   Source `gorm:"foreignKey:SourceID"`
	// DedupHash is a fixed-width digest for ingest deduplication (e.g. Reddit: SHA-256 hex of post.ID); nil for rows with no dedupe key.
	DedupHash *string `gorm:"size:64;uniqueIndex:idx_intel_source_dedup"`
	// Content is canonical markdown for display and VL text input.
	Content string `gorm:"type:text;not null;default:''"`
	// PreviewImageID optionally points at a copied thumbnail in [p_filesystem] (see [persistRedditPreviewImage]).
	PreviewImageID *uint
	PreviewImage   p_filesystem.VNode `gorm:"foreignKey:PreviewImageID"`
	// Embedding is pgvector storage; dimension must match [IntelEmbeddingDim]. Required (NOT NULL); starts as zeros until [applyIntelEmbedding] runs when a [VLEmbedder] is registered.
	Embedding pgvector.Vector `gorm:"type:vector(1024);not null"`
}

// BeforeCreate sets a zero placeholder so the first INSERT satisfies NOT NULL; [Intel.AfterSave] replaces it when a VL embedder is configured.
func (i *Intel) BeforeCreate(tx *gorm.DB) error {
	if len(i.Embedding.Slice()) != IntelEmbeddingDim {
		i.Embedding = pgvector.NewVector(make([]float32, IntelEmbeddingDim))
	}
	return nil
}

// AfterSave refreshes [Intel.Embedding] on every create/update (embedding write uses SkipHooks to avoid recursion).
func (i *Intel) AfterSave(tx *gorm.DB) error {
	if tx.Statement.SkipHooks {
		return nil
	}
	applyIntelEmbedding(context.Background(), tx, i)
	return nil
}

func init() {
	lago.OnDBInit("p_lacerate.intel_model", func(db *gorm.DB) *gorm.DB {
		backfillIntelNullContentBeforeMigrate(db)
		backfillIntelNullEmbeddingBeforeMigrate(db)
		lago.RegisterModel[Intel](db)
		return db
	})
}
