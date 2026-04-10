package p_lacerate

import (
	"context"
	"log/slog"
	"time"

	"github.com/lariv-in/lago/lago"
	"github.com/pgvector/pgvector-go"
	"gorm.io/gorm"
)

// Lookup stores a single text payload (e.g. reference or canned content).
type Lookup struct {
	gorm.Model
	Content string `gorm:"type:text;not null;default:''"`
	// UpdateInterval, when non-nil and positive, schedules the OSINT lookup agent ([runLookupWorker]).
	UpdateInterval *time.Duration
	// Embedding matches [IntelEmbeddingDim] and [Lookup.Content]; NOT NULL with zero prefill ([Lookup.BeforeCreate]); refreshed on save ([applyLookupEmbedding]).
	Embedding pgvector.Vector `gorm:"type:vector(1024);not null"`
}

// BeforeCreate sets a zero placeholder so the first INSERT satisfies NOT NULL; [Lookup.AfterSave] replaces it when a [VLEmbedder] is registered.
func (l *Lookup) BeforeCreate(tx *gorm.DB) error {
	if len(l.Embedding.Slice()) != IntelEmbeddingDim {
		l.Embedding = pgvector.NewVector(make([]float32, IntelEmbeddingDim))
	}
	return nil
}

// AfterSave refreshes [Lookup.Embedding] on every create/update (embedding write uses SkipHooks to avoid recursion).
func (l *Lookup) AfterSave(tx *gorm.DB) error {
	if tx.Statement.SkipHooks {
		return nil
	}
	applyLookupEmbedding(context.Background(), tx, l)
	// Defer worker restart until after the surrounding transaction commits ([views.LayerCreate] / [views.LayerUpdate] use a transaction).
	scheduleRestartLookupWorker(tx, l.ID)
	return nil
}

// AfterDelete stops the background worker for this lookup.
func (l *Lookup) AfterDelete(tx *gorm.DB) error {
	if l != nil && l.ID != 0 {
		StopLookupWorker(l.ID)
	}
	return nil
}

func backfillLookupNullEmbeddingBeforeMigrate(db *gorm.DB) {
	if !db.Migrator().HasTable("lookups") {
		return
	}
	if !db.Migrator().HasColumn(&Lookup{}, "Embedding") {
		return
	}
	if db.Dialector.Name() != "postgres" {
		return
	}
	lit := intelZeroVectorTextForPG()
	if err := db.Exec(`UPDATE lookups SET embedding = ?::vector WHERE embedding IS NULL`, lit).Error; err != nil {
		slog.Error("lacerate: backfill lookup null embedding before migrate", "error", err)
	}
}

// backfillLookupZeroIntervalToNull clears zero/negative stored durations after the column is nullable.
func backfillLookupZeroIntervalToNull(db *gorm.DB) {
	if !db.Migrator().HasTable("lookups") {
		return
	}
	if !db.Migrator().HasColumn(&Lookup{}, "UpdateInterval") {
		return
	}
	if err := db.Exec(`UPDATE lookups SET update_interval = NULL WHERE update_interval IS NOT NULL AND update_interval <= 0`).Error; err != nil {
		slog.Error("lacerate: backfill lookup zero interval to null", "error", err)
	}
}

func init() {
	lago.OnDBInit("p_lacerate.lookup_model", func(db *gorm.DB) *gorm.DB {
		backfillLookupNullEmbeddingBeforeMigrate(db)
		lago.RegisterModel[Lookup](db)
		backfillLookupZeroIntervalToNull(db)
		startLookupWorkersFromDB(db)
		return db
	})
}
