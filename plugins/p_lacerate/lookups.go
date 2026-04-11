package p_lacerate

import (
	"context"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"gorm.io/gorm"
)

type lookupWorkerPhase uint32

const (
	lookupWorkerPhaseWaiting lookupWorkerPhase = iota
	lookupWorkerPhaseRunning
)

type lookupWorkerHandle struct {
	ctx    context.Context
	cancel context.CancelFunc
	phase  atomic.Uint32
}

var (
	lookupWorkerMu sync.Mutex
	lookupWorkers  = map[uint]*lookupWorkerHandle{}
)

// RunLookupNow restarts the background worker when [Lookup.UpdateInterval] is positive; otherwise runs
// one [runLookupAgent] in a new goroutine (no scheduled worker).
func RunLookupNow(db *gorm.DB, lookupID uint) {
	if db == nil || lookupID == 0 {
		return
	}
	var lu Lookup
	if err := db.First(&lu, lookupID).Error; err != nil {
		slog.Error("lacerate: RunLookupNow load lookup", "error", err, "lookup_id", lookupID)
		return
	}
	if lu.UpdateInterval != nil && *lu.UpdateInterval > 0 {
		ScheduleRestartLookupWorker(db, &lu)
		return
	}
	go func() {
		ctx := context.Background()
		var fresh Lookup
		if err := db.WithContext(ctx).First(&fresh, lookupID).Error; err != nil {
			slog.Error("lacerate: RunLookupNow one-shot reload", "error", err, "lookup_id", lookupID)
			return
		}
		if err := runLookupAgent(ctx, db.WithContext(ctx), &fresh); err != nil {
			slog.Error("lacerate: RunLookupNow one-shot agent", "error", err, "lookup_id", lookupID)
		}
	}()
}

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

func lookupWorkerSetPhase(lookupID uint, p lookupWorkerPhase) {
	lookupWorkerMu.Lock()
	h := lookupWorkers[lookupID]
	lookupWorkerMu.Unlock()
	if h == nil {
		return
	}
	h.phase.Store(uint32(p))
}

// LookupWorkerIsRunning reports whether a scheduled background worker is registered for this lookup
// (including sleep between runs). False when [Lookup.UpdateInterval] is unset/non-positive or worker exited.
func LookupWorkerIsRunning(lookupID uint) bool {
	if lookupID == 0 {
		return false
	}
	lookupWorkerMu.Lock()
	defer lookupWorkerMu.Unlock()
	_, ok := lookupWorkers[lookupID]
	return ok
}

// LookupWorkerRunning reports whether the lookup worker is running the agent (true) or sleeping (false).
// ok is false when no worker is registered for lookupID.
func LookupWorkerRunning(lookupID uint) (running bool, ok bool) {
	if lookupID == 0 {
		return false, false
	}
	lookupWorkerMu.Lock()
	h := lookupWorkers[lookupID]
	lookupWorkerMu.Unlock()
	if h == nil {
		return false, false
	}
	return lookupWorkerPhase(h.phase.Load()) == lookupWorkerPhaseRunning, true
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
	h := &lookupWorkerHandle{ctx: ctx, cancel: cancel}
	h.phase.Store(uint32(lookupWorkerPhaseRunning))
	lookupWorkerMu.Lock()
	lookupWorkers[lu.ID] = h
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

		lookupWorkerSetPhase(lookupID, lookupWorkerPhaseRunning)

		if lu.UpdateInterval == nil || *lu.UpdateInterval <= 0 {
			StopLookupWorker(lookupID)
			return
		}

		if err := runLookupAgent(ctx, db.WithContext(ctx), &lu); err != nil {
			slog.Error("lacerate: lookup agent run", "error", err, "lookup_id", lookupID)
		}

		lookupWorkerSetPhase(lookupID, lookupWorkerPhaseWaiting)
		interval := *lu.UpdateInterval
		select {
		case <-ctx.Done():
			return
		case <-time.After(interval):
		}
	}
}
