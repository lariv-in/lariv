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
)

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

// scheduleRestartSourceWorker restarts the source fetch worker without blocking [Source.AfterSave].
func scheduleRestartSourceWorker(tx *gorm.DB, sourceID uint) {
	if sourceID == 0 || tx == nil {
		return
	}
	id := sourceID
	sess := tx.Session(&gorm.Session{NewDB: true})
	go func() {
		RestartSourceWorker(sess, id)
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
	_, err := row.Fetch(ctx, db.WithContext(ctx))
	return err
}
