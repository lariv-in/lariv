package p_seer_websites

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/lariv-in/lago/syncmap"
	"gorm.io/gorm"
)

var websiteRunnerWorkerPoolCancels *syncmap.SyncMap[uint, context.CancelFunc] = &syncmap.SyncMap[uint, context.CancelFunc]{}

// WebsiteRunnerWorkerPoolIsRunning reports whether a worker-pool goroutine is registered for runnerID.
func WebsiteRunnerWorkerPoolIsRunning(runnerID uint) bool {
	if runnerID == 0 {
		return false
	}
	cancel, ok := websiteRunnerWorkerPoolCancels.Load(runnerID)
	return ok && cancel != nil
}

// StopWebsiteRunnerWorkerPool cancels the pool goroutine for runnerID and removes it from the map.
func StopWebsiteRunnerWorkerPool(runnerID uint) {
	if runnerID == 0 {
		return
	}
	cancel, loaded := websiteRunnerWorkerPoolCancels.LoadAndDelete(runnerID)
	if loaded && cancel != nil {
		slog.Info("p_seer_websites: worker pool stop requested", "runner_id", runnerID)
		cancel()
	}
}

// ScheduleWebsiteRunnerWorkerPoolStart starts the pool in a new goroutine if not already running.
func ScheduleWebsiteRunnerWorkerPoolStart(db *gorm.DB, runnerID uint) {
	if db == nil || runnerID == 0 {
		return
	}
	d := db
	id := runnerID
	go func() {
		ctx, cancel := context.WithCancel(context.Background())
		if _, loaded := websiteRunnerWorkerPoolCancels.LoadOrStore(id, cancel); loaded {
			cancel()
			slog.Info("p_seer_websites: worker pool start skipped (already running)", "runner_id", id)
			return
		}

		slog.Info("p_seer_websites: worker pool goroutine scheduled", "runner_id", id)
		runWebsiteRunnerWorkerPool(d.WithContext(ctx), id, ctx)
	}()
}

func runWebsiteRunnerWorkerPool(db *gorm.DB, runnerID uint, ctx context.Context) {
	defer slog.Info("p_seer_websites: worker pool exited", "runner_id", runnerID)

	var loggedRunnerMeta bool
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		var runner WebsiteRunner
		if err := db.First(&runner, runnerID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				slog.Warn("p_seer_websites: worker pool runner row missing", "runner_id", runnerID)
			} else {
				slog.Error("p_seer_websites: worker pool load runner", "error", err, "runner_id", runnerID)
			}
			StopWebsiteRunnerWorkerPool(runnerID)
			return
		}

		if !loggedRunnerMeta {
			slog.Info("p_seer_websites: worker pool active",
				"runner_id", runnerID,
				"runner_name", runner.Name,
				"interval", runner.Duration.String(),
			)
			loggedRunnerMeta = true
		}

		var sources []WebsiteSource
		if err := db.Where("website_runner_id = ?", runnerID).Find(&sources).Error; err != nil {
			slog.Error("p_seer_websites: worker pool list sources", "error", err, "runner_id", runnerID)
		} else {
			slog.Info("p_seer_websites: worker pool pass",
				"runner_id", runnerID,
				"runner_name", runner.Name,
				"source_count", len(sources),
			)
			for i := range sources {
				src := sources[i]
				seed := strings.TrimSpace(src.URL.String())
				start := time.Now()
				slog.Info("p_seer_websites: worker fetch start",
					"runner_id", runnerID,
					"website_source_id", src.ID,
					"seed_url", seed,
					"depth", src.Depth,
				)
				if err := src.Fetch(ctx, db); err != nil {
					slog.Error("p_seer_websites: worker fetch",
						"error", err,
						"runner_id", runnerID,
						"website_source_id", src.ID,
						"elapsed", time.Since(start),
					)
				} else {
					slog.Info("p_seer_websites: worker fetch ok",
						"runner_id", runnerID,
						"website_source_id", src.ID,
						"elapsed", time.Since(start),
					)
				}
			}
		}

		if runner.Duration <= 0 {
			slog.Info("p_seer_websites: worker pool stopping (interval not positive)", "runner_id", runnerID)
			StopWebsiteRunnerWorkerPool(runnerID)
			return
		}

		slog.Debug("p_seer_websites: worker pool sleeping",
			"runner_id", runnerID,
			"duration", runner.Duration.String(),
		)

		select {
		case <-ctx.Done():
			return
		case <-time.After(runner.Duration):
		}
	}
}
