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
	if _, ok := SourceKindMap[src.Kind]; !ok {
		err := fmt.Errorf("unknown source kind %q", src.Kind)
		slog.Error("lacerate: source fetch", "error", err, "source_id", src.ID, "kind", src.Kind)
		return err
	}

	switch src.Kind {
	case "reddit":
		var rs RedditSource
		if err := db.Preload("Source").Where("source_id = ?", src.ID).First(&rs).Error; err != nil {
			slog.Error("lacerate: source fetch load reddit source", "error", err, "source_id", src.ID)
			return err
		}
		_, err := rs.Fetch(ctx, db.WithContext(ctx))
		return err
	case "twitter":
		var ts TwitterSource
		if err := db.Preload("Source").Where("source_id = ?", src.ID).First(&ts).Error; err != nil {
			slog.Error("lacerate: source fetch load twitter source", "error", err, "source_id", src.ID)
			return err
		}
		_, err := ts.Fetch(ctx, db.WithContext(ctx))
		return err
	default:
		err := fmt.Errorf("source worker not implemented for kind %q", src.Kind)
		slog.Error("lacerate: source fetch", "error", err, "source_id", src.ID, "kind", src.Kind)
		return err
	}
}
