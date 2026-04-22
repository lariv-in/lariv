package p_seer_deepsearch

import (
	"context"

	"github.com/lariv-in/lago/syncmap"
	"gorm.io/gorm"
)

// deepSearchStopCancels holds user-cancel functions for in-flight pipelines (see [BeginDeepSearchPipeline]).
var deepSearchStopCancels = &syncmap.SyncMap[uint, context.CancelFunc]{}

// BeginDeepSearchPipeline registers a stop cancel func then runs [runDeepSearchPipeline] in a background goroutine.
func BeginDeepSearchPipeline(db *gorm.DB, id uint) {
	ctx, stopCancel := context.WithCancel(context.Background())
	deepSearchStopCancels.Store(id, stopCancel)
	dbCopy := db
	go func() {
		defer deepSearchStopCancels.Delete(id)
		defer stopCancel()
		runDeepSearchPipeline(ctx, dbCopy, id)
	}()
}

// TryStopDeepSearchPipeline invokes the registered cancel for [id], if any. Returns whether a cancel was invoked.
func TryStopDeepSearchPipeline(id uint) bool {
	fn, ok := deepSearchStopCancels.Load(id)
	if !ok {
		return false
	}
	fn()
	return true
}

func deepSearchFinishCancelled(ctx context.Context, db *gorm.DB, id uint) {
	appendDeepSearchLog(ctx, db, id, DeepSearchLogKindInfo, "pipeline stopped (cancel requested)")
	_ = persistDeepSearch(ctx, db, id, map[string]any{
		"status":    DeepSearchStatusCancelled,
		"run_error": "",
	})
}

// deepSearchAbortIfCtxDone persists terminal status when ctx is done. Returns true if the pipeline should exit.
func deepSearchAbortIfCtxDone(ctx context.Context, db *gorm.DB, id uint) bool {
	err := ctx.Err()
	if err == nil {
		return false
	}
	deepSearchFinishCancelled(ctx, db, id)
	return true
}
