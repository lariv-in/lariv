package p_lacerate

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/lariv-in/lago/lago"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// intelCreateBatchSize is chunk size for [runSourceFetch] batch insert after embeddings.
const intelCreateBatchSize = 100

// sourceWorkerHandle holds a cancellable context for one [Source] worker goroutine.
type sourceWorkerHandle struct {
	ctx    context.Context
	cancel context.CancelFunc
}

var (
	sourceWorkerMu sync.Mutex
	sourceWorkers  = map[uint]sourceWorkerHandle{}
)

func init() {
	lago.OnDBInit("p_lacerate.sources_workers", func(db *gorm.DB) *gorm.DB {
		var sources []Source
		if err := db.Find(&sources).Error; err != nil {
			slog.Error("lacerate: load sources for workers", "error", err)
			return db
		}
		for i := range sources {
			RestartSourceWorker(db, sources[i].ID)
		}
		return db
	})
}

// ScheduleRestartSourceWorker restarts the source fetch worker in a goroutine without blocking the caller.
// db must be a pooled *gorm.DB (e.g. request DB from context), not a transactional *gorm.DB from inside db.Transaction;
// call after the transaction returns so the worker does not share the transaction connection (avoids "conn busy").
func ScheduleRestartSourceWorker(db *gorm.DB, sourceID uint) {
	if sourceID == 0 || db == nil {
		return
	}
	id := sourceID
	go func() {
		RestartSourceWorker(db, id)
	}()
}

// StopSourceWorker cancels the worker for this source ID and removes it from the map.
func StopSourceWorker(sourceID uint) {
	sourceWorkerMu.Lock()
	defer sourceWorkerMu.Unlock()
	if w, ok := sourceWorkers[sourceID]; ok {
		w.cancel()
		delete(sourceWorkers, sourceID)
	}
}

// RestartSourceWorker stops an existing worker (if any), then starts a new one using the
// current row from the DB. No worker is started if the source is missing or [Source.Duration] <= 0.
func RestartSourceWorker(db *gorm.DB, sourceID uint) {
	if db == nil {
		slog.Error("lacerate: RestartSourceWorker called with nil db", "source_id", sourceID)
		return
	}

	StopSourceWorker(sourceID)

	var src Source
	if err := db.First(&src, sourceID).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			slog.Error("lacerate: load source for worker", "error", err, "source_id", sourceID)
		}
		return
	}
	if src.Duration <= 0 {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	sourceWorkerMu.Lock()
	sourceWorkers[sourceID] = sourceWorkerHandle{ctx: ctx, cancel: cancel}
	sourceWorkerMu.Unlock()

	go runSourceWorker(db, sourceID, ctx)
}

func runSourceWorker(db *gorm.DB, sourceID uint, ctx context.Context) {
	defer slog.Info("lacerate: source worker exited", "source_id", sourceID)

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		var src Source
		if err := db.WithContext(ctx).First(&src, sourceID).Error; err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				slog.Error("lacerate: source worker reload source", "error", err, "source_id", sourceID)
			}
			return
		}
		if src.Duration <= 0 {
			StopSourceWorker(sourceID)
			return
		}

		if err := runSourceFetch(ctx, db, &src); err != nil {
			slog.Error("lacerate: source worker fetch", "error", err, "source_id", sourceID, "kind", src.Kind)
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(src.Duration):
		}
	}
}

func runSourceFetch(ctx context.Context, db *gorm.DB, src *Source) error {
	newRow, ok := RegistrySourceKind.Get(src.Kind)
	if !ok {
		err := fmt.Errorf("source worker not registered for kind %q", src.Kind)
		slog.Error("lacerate: source fetch", "error", err, "source_id", src.ID, "kind", src.Kind)
		return err
	}
	row := newRow()
	if err := db.Preload("Source").Model(row).Where("source_id = ?", src.ID).First(row).Error; err != nil {
		slog.Error("lacerate: source fetch load kind row", "error", err, "source_id", src.ID, "kind", src.Kind)
		return err
	}

	dbw := db.WithContext(ctx)
	var hashRows []string
	if err := dbw.Model(&Intel{}).Where("source_id = ? AND dedup_hash IS NOT NULL AND dedup_hash <> ''", src.ID).Pluck("dedup_hash", &hashRows).Error; err != nil {
		slog.Error("lacerate: source fetch pluck dedup hashes", "error", err, "source_id", src.ID)
		return err
	}
	existingDedup := make(map[string]struct{}, len(hashRows))
	for _, h := range hashRows {
		existingDedup[h] = struct{}{}
	}

	intels, err := row.Fetch(ctx, dbw, existingDedup)
	if err != nil {
		return err
	}
	if len(intels) == 0 {
		return nil
	}

	toSave := make([]Intel, 0, len(intels))
	for i := range intels {
		dh := intels[i].DedupHash
		if dh == nil || *dh == "" {
			continue
		}
		toSave = append(toSave, intels[i])
	}
	if len(toSave) == 0 {
		return nil
	}

	for i := range toSave {
		if err := prepareIntelEmbeddingForSave(ctx, dbw, &toSave[i]); err != nil {
			return err
		}
	}

	intelOnConflict := clause.OnConflict{
		Columns: []clause.Column{
			{Name: "source_id"},
			{Name: "dedup_hash"},
		},
		DoNothing: true,
	}
	if err := dbw.Session(&gorm.Session{SkipHooks: true}).
		Clauses(intelOnConflict).
		CreateInBatches(toSave, intelCreateBatchSize).Error; err != nil {
		slog.Error("lacerate: source fetch batch create intel", "error", err, "source_id", src.ID)
		return err
	}
	return nil
}
