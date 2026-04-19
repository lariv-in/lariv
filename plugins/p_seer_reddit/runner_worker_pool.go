package p_seer_reddit

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/lariv-in/lago/syncmap"
	"gorm.io/gorm"
)

// redditRunnerWorkerPool is one cooperative goroutine per [RedditRunner]: it loads sources
// with this runner id, runs [FetchNewRedditPosts] on each, then sleeps [RedditRunner.Duration]
// before the next pass (until stopped or runner row missing).
// Map values are the pool's [context.CancelFunc] (nil means slot unused; callers should not store nil).
var redditRunnerWorkerPoolCancels *syncmap.SyncMap[uint, context.CancelFunc] = &syncmap.SyncMap[uint, context.CancelFunc]{}

// RedditRunnerWorkerPoolIsRunning reports whether a worker-pool goroutine is registered for runnerID.
func RedditRunnerWorkerPoolIsRunning(runnerID uint) bool {
	if runnerID == 0 {
		return false
	}
	cancel, ok := redditRunnerWorkerPoolCancels.Load(runnerID)
	return ok && cancel != nil
}

// StopRedditRunnerWorkerPool cancels the pool goroutine for runnerID and removes it from the map.
func StopRedditRunnerWorkerPool(runnerID uint) {
	if runnerID == 0 {
		return
	}
	cancel, loaded := redditRunnerWorkerPoolCancels.LoadAndDelete(runnerID)
	if loaded && cancel != nil {
		cancel()
	}
}

// ScheduleRedditRunnerWorkerPoolStart starts the pool in a new goroutine if not already running.
// db must be a pooled *gorm.DB (not an open transaction).
func ScheduleRedditRunnerWorkerPoolStart(db *gorm.DB, runnerID uint) {
	if db == nil || runnerID == 0 {
		return
	}
	d := db
	id := runnerID
	go func() {
		ctx, cancel := context.WithCancel(context.Background())
		if _, loaded := redditRunnerWorkerPoolCancels.LoadOrStore(id, cancel); loaded {
			cancel()
			return
		}

		runRedditRunnerWorkerPool(d.WithContext(ctx), id, ctx)
	}()
}

func runRedditRunnerWorkerPool(db *gorm.DB, runnerID uint, ctx context.Context) {
	defer slog.Info("p_seer_reddit: worker pool exited", "runner_id", runnerID)

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		var runner RedditRunner
		if err := db.First(&runner, runnerID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				slog.Warn("p_seer_reddit: worker pool runner row missing", "runner_id", runnerID)
			} else {
				slog.Error("p_seer_reddit: worker pool load runner", "error", err, "runner_id", runnerID)
			}
			StopRedditRunnerWorkerPool(runnerID)
			return
		}

		var sources []RedditSource
		if err := db.Where("reddit_runner_id = ?", runnerID).Find(&sources).Error; err != nil {
			slog.Error("p_seer_reddit: worker pool list sources", "error", err, "runner_id", runnerID)
		} else {
			for i := range sources {
				src := sources[i]
				if err := FetchNewRedditPosts(ctx, db, &src); err != nil {
					slog.Error("p_seer_reddit: worker pool fetch",
						"error", err,
						"runner_id", runnerID,
						"reddit_source_id", src.ID,
					)
				}
			}
		}

		if runner.Duration <= 0 {
			StopRedditRunnerWorkerPool(runnerID)
			return
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(runner.Duration):
		}
	}
}
