package p_lacerate

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

	"gorm.io/gorm"
)

type lookupWorkerHandle struct {
	ctx    context.Context
	cancel context.CancelFunc
}

var (
	lookupWorkerMu sync.Mutex
	lookupWorkers  = map[uint]lookupWorkerHandle{}
)

// scheduleRestartLookupWorker restarts the interval worker after the surrounding transaction commits.
func scheduleRestartLookupWorker(tx *gorm.DB, lookupID uint) {
	if lookupID == 0 || tx == nil {
		return
	}
	sess := tx.Session(&gorm.Session{NewDB: true})
	go func() {
		time.Sleep(40 * time.Millisecond)
		RestartLookupWorker(sess, lookupID)
	}()
}

// startLookupWorkersFromDB starts workers for every lookup that has a positive update interval.
func startLookupWorkersFromDB(db *gorm.DB) {
	if db == nil {
		slog.Error("lacerate: startLookupWorkersFromDB nil db")
		return
	}
	var lookups []Lookup
	if err := db.Where("update_interval IS NOT NULL").Where("update_interval > ?", 0).Find(&lookups).Error; err != nil {
		slog.Error("lacerate: load lookups for workers", "error", err)
		return
	}
	for i := range lookups {
		RestartLookupWorker(db, lookups[i].ID)
	}
}

// StopLookupWorker cancels the worker for this lookup ID and removes it from the map.
func StopLookupWorker(lookupID uint) {
	lookupWorkerMu.Lock()
	defer lookupWorkerMu.Unlock()
	if w, ok := lookupWorkers[lookupID]; ok {
		w.cancel()
		delete(lookupWorkers, lookupID)
	}
}

// RestartLookupWorker stops an existing worker (if any), then starts a new one using the
// current row from the DB. No worker is started if the lookup is missing or the interval is nil/non-positive.
func RestartLookupWorker(db *gorm.DB, lookupID uint) {
	if db == nil {
		slog.Error("lacerate: RestartLookupWorker called with nil db", "lookup_id", lookupID)
		return
	}

	StopLookupWorker(lookupID)

	var lu Lookup
	if err := db.First(&lu, lookupID).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			slog.Error("lacerate: load lookup for worker", "error", err, "lookup_id", lookupID)
		}
		return
	}
	if lu.UpdateInterval == nil || *lu.UpdateInterval <= 0 {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	lookupWorkerMu.Lock()
	lookupWorkers[lookupID] = lookupWorkerHandle{ctx: ctx, cancel: cancel}
	lookupWorkerMu.Unlock()

	go runLookupWorker(db, lookupID, ctx)
}

func runLookupWorker(db *gorm.DB, lookupID uint, ctx context.Context) {
	defer slog.Info("lacerate: lookup worker exited", "lookup_id", lookupID)

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		var lu Lookup
		if err := db.WithContext(ctx).First(&lu, lookupID).Error; err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				slog.Error("lacerate: lookup worker reload lookup", "error", err, "lookup_id", lookupID)
			}
			return
		}
		if lu.UpdateInterval == nil || *lu.UpdateInterval <= 0 {
			StopLookupWorker(lookupID)
			return
		}

		if err := runLookupAgent(ctx, db.WithContext(ctx), &lu); err != nil {
			slog.Error("lacerate: lookup agent run", "error", err, "lookup_id", lookupID)
		}

		interval := *lu.UpdateInterval
		select {
		case <-ctx.Done():
			return
		case <-time.After(interval):
		}
	}
}
