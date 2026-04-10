package p_lacerate

import (
	"context"
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

// ScheduleRestartLookupWorker restarts the lookup worker in a goroutine without blocking the caller.
// db must be a pooled *gorm.DB (e.g. after db.Transaction returns), not a transactional *gorm.DB.
func ScheduleRestartLookupWorker(db *gorm.DB, lu *Lookup) {
	if lu == nil || lu.ID == 0 || db == nil {
		return
	}
	luc := *lu
	go func() {
		RestartLookupWorker(db, &luc)
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
		RestartLookupWorker(db, &lookups[i])
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

// RestartLookupWorker stops an existing worker (if any), then starts a new one from lu.
// No worker is started if lu is nil, lu.ID is zero, or the interval is nil/non-positive.
// The worker uses this snapshot until the next restart (e.g. after a save).
func RestartLookupWorker(db *gorm.DB, lu *Lookup) {
	if db == nil {
		slog.Error("lacerate: RestartLookupWorker called with nil db")
		return
	}
	if lu == nil || lu.ID == 0 {
		return
	}

	StopLookupWorker(lu.ID)

	if lu.UpdateInterval == nil || *lu.UpdateInterval <= 0 {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	lookupWorkerMu.Lock()
	lookupWorkers[lu.ID] = lookupWorkerHandle{ctx: ctx, cancel: cancel}
	lookupWorkerMu.Unlock()

	go runLookupWorker(db, *lu, ctx)
}

func runLookupWorker(db *gorm.DB, lu Lookup, ctx context.Context) {
	lookupID := lu.ID
	defer slog.Info("lacerate: lookup worker exited", "lookup_id", lookupID)

	for {
		select {
		case <-ctx.Done():
			return
		default:
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
