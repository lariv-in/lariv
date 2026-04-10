package p_lacerate

import (
	"context"
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
	// Embedding matches [IntelEmbeddingDim] and [Lookup.Content]; NOT NULL with zero prefill ([Lookup.BeforeCreate]); set before write ([prepareLookupEmbeddingForSave]).
	Embedding pgvector.Vector `gorm:"type:vector(1024);not null"`
}

// BeforeCreate sets a zero placeholder so the first INSERT satisfies NOT NULL when [prepareLookupEmbeddingForSave] does not run (e.g. no [VLEmbedder] yet).
func (l *Lookup) BeforeCreate(tx *gorm.DB) error {
	if len(l.Embedding.Slice()) != IntelEmbeddingDim {
		l.Embedding = pgvector.NewVector(make([]float32, IntelEmbeddingDim))
	}
	return nil
}

// BeforeSave fills [Lookup.Embedding] from [Lookup.Content] so the same INSERT/UPDATE persists the vector.
func (l *Lookup) BeforeSave(tx *gorm.DB) error {
	if tx.Statement.SkipHooks {
		return nil
	}
	return prepareLookupEmbeddingForSave(context.Background(), l)
}

// AfterTxCommit implements [views.TxCommitHook] so the lookup worker restarts after LayerCreate/LayerUpdate
// commit using the pooled DB (avoids "conn busy" from scheduling inside the transaction).
func (l *Lookup) AfterTxCommit(db *gorm.DB) {
	if l == nil {
		return
	}
	ScheduleRestartLookupWorker(db, l)
}

// AfterDelete stops the background worker for this lookup.
func (l *Lookup) AfterDelete(tx *gorm.DB) error {
	if l != nil && l.ID != 0 {
		StopLookupWorker(l.ID)
	}
	return nil
}

func init() {
	lago.OnDBInit("p_lacerate.lookup_model", func(db *gorm.DB) *gorm.DB {
		lago.RegisterModel[Lookup](db)
		startLookupWorkersFromDB(db)
		return db
	})
}
