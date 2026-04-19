package p_seer_reddit

import (
	"context"
	"fmt"
	"log/slog"
	"sync/atomic"
	"time"

	"github.com/lariv-in/lago/plugins/p_seer_intel"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"
)

// Global intel ingest gate: at most one Reddit→Intel job at a time ([redditIntelIngestActive]).
var redditIntelIngestActive atomic.Bool

// bulkIntelParallelism caps concurrent GenAI + DB work per bulk ingest run.
const bulkIntelParallelism = 8

// createIntelForRedditPostIfMissing creates an [p_seer_intel.Intel] row when none exists yet for the post.
// Returns nil when skipped (already exists) or on success; returns a wrapped error on GenAI or DB failure.
func createIntelForRedditPostIfMissing(ctx context.Context, db *gorm.DB, post RedditPost) error {
	kind := (RedditPost{}).Kind()
	exists, err := p_seer_intel.IntelExistsForSource(ctx, db, kind, post.ID)
	if err != nil {
		return fmt.Errorf("exists check: %w", err)
	}
	if exists {
		return nil
	}
	intel, err := p_seer_intel.NewFromIntelKind(ctx, &post)
	if err != nil {
		return fmt.Errorf("generate: %w", err)
	}
	if err := db.WithContext(ctx).Create(&intel).Error; err != nil {
		return fmt.Errorf("persist: %w", err)
	}
	return nil
}

// RunRedditBulkIntelIngest runs [createIntelForRedditPostIfMissing] for each post with bounded parallelism.
func RunRedditBulkIntelIngest(ctx context.Context, db *gorm.DB, posts []RedditPost) {
	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(bulkIntelParallelism)
	for i := range posts {
		post := posts[i]
		if post.ID == 0 {
			continue
		}
		p := post
		g.Go(func() error {
			if err := createIntelForRedditPostIfMissing(ctx, db, p); err != nil {
				slog.Warn("p_seer_reddit: bulk add intel post", "post_id", p.ID, "error", err)
			}
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		slog.Warn("p_seer_reddit: bulk add intel group", "error", err)
	}
}

// RunRedditSinglePostIntelIngest runs [createIntelForRedditPostIfMissing] for one post in the background window.
func RunRedditSinglePostIntelIngest(ctx context.Context, db *gorm.DB, post RedditPost) {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Minute)
	defer cancel()
	if err := createIntelForRedditPostIfMissing(ctx, db, post); err != nil {
		slog.Warn("p_seer_reddit: single add intel post", "post_id", post.ID, "error", err)
	}
}
