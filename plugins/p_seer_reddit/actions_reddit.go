package p_seer_reddit

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"
)

const defaultMaxFreshPosts = 25

// redditFetchPostsActive is true while a source "Load" ([FetchNewRedditPosts]) runs in the background.
var redditFetchPostsActive atomic.Bool

// FetchNewRedditPosts loads Reddit JSON listings and inserts or links [RedditPost] rows only.
// Call this from HTTP handlers, jobs, or orchestration outside Runnable/worker wiring.
func FetchNewRedditPosts(ctx context.Context, db *gorm.DB, src *RedditSource) error {
	if src == nil || src.ID == 0 {
		return fmt.Errorf("p_seer_reddit: reddit source not loaded")
	}
	var subs []string
	if len(src.Subreddits) > 0 {
		if err := json.Unmarshal(src.Subreddits, &subs); err != nil {
			slog.Error("p_seer_reddit: unmarshal subreddits", "error", err, "reddit_source_id", src.ID)
			return err
		}
	}
	query := strings.TrimSpace(src.SearchQuery)
	maxFresh := int(src.MaxFreshPosts)
	if maxFresh == 0 {
		maxFresh = defaultMaxFreshPosts
	}

	var names []string
	for _, raw := range subs {
		name := strings.TrimSpace(strings.TrimPrefix(raw, "r/"))
		if name != "" {
			names = append(names, name)
		}
	}
	if len(names) == 0 {
		return nil
	}

	var mu sync.Mutex
	var freshTotal int
	g, gctx := errgroup.WithContext(ctx)
	for _, name := range names {
		g.Go(func() error {
			return src.fetchSubredditListings(gctx, db, name, query, maxFresh, &mu, &freshTotal)
		})
	}
	return g.Wait()
}

func (r *RedditSource) fetchSubredditListings(ctx context.Context, db *gorm.DB, subredditName, searchQuery string, maxFresh int, mu *sync.Mutex, freshTotal *int) error {
	afterStr := ""
	for range redditListingMaxPages {
		if err := ctx.Err(); err != nil {
			return err
		}
		mu.Lock()
		atCap := *freshTotal >= maxFresh
		mu.Unlock()
		if atCap {
			return nil
		}
		var afterPtr *string
		if afterStr != "" {
			afterPtr = &afterStr
		}
		var listing *redditObject[redditListing[RedditPostData]]
		var err error
		if searchQuery != "" {
			listing, err = fetchSubredditPostsSearch(ctx, subredditName, searchQuery, afterPtr)
		} else {
			listing, err = fetchSubredditPosts(ctx, subredditName, afterPtr)
		}
		if err != nil {
			return err
		}
		for _, child := range listing.Data.Children {
			post := child.Data
			id := strings.TrimSpace(post.ID)
			if id == "" {
				slog.Warn("p_seer_reddit: post missing id, skip", "subreddit", subredditName)
				continue
			}
			if post.CreatedUTC <= 0 {
				err := fmt.Errorf("reddit post %q missing datetime", post.ID)
				slog.Error("p_seer_reddit: post datetime", "error", err, "reddit_source_id", r.ID)
				return err
			}
			mu.Lock()
			if *freshTotal >= maxFresh {
				mu.Unlock()
				return nil
			}
			inserted, err := r.persistPostIfNew(ctx, db, post)
			if err != nil {
				mu.Unlock()
				return err
			}
			if inserted {
				*freshTotal++
			}
			mu.Unlock()
		}
		if listing.Data.After == nil {
			break
		}
		afterStr = *listing.Data.After
	}
	return nil
}

func redditSourcePostLinkExists(ctx context.Context, db *gorm.DB, sourceID uint, postID string) (bool, error) {
	var n int64
	err := db.WithContext(ctx).Raw(
		`SELECT COUNT(*) FROM `+RedditSourcePostsJoinTable+` rsp
		 INNER JOIN `+RedditPostsTable+` p ON p.id = rsp.reddit_post_id
		 WHERE rsp.reddit_source_id = ? AND p.post_id = ? AND p.deleted_at IS NULL`,
		sourceID, postID,
	).Scan(&n).Error
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

// ErrRedditPostSoftDeleteNotFound is returned when no active row matches the id for [SoftWipeRedditPost].
var ErrRedditPostSoftDeleteNotFound = errors.New("p_seer_reddit: reddit post not found or already deleted")

// SoftWipeRedditPost sets [gorm.Model.DeletedAt] and clears all stored fields except [RedditPost.PostID].
func SoftWipeRedditPost(ctx context.Context, db *gorm.DB, id uint) error {
	if id == 0 {
		return fmt.Errorf("p_seer_reddit: invalid post id")
	}
	now := time.Now().UTC()
	tx := db.WithContext(ctx).Model(&RedditPost{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Select("DeletedAt", "Title", "Selftext", "Author", "Subreddit", "Permalink", "URL", "CreatedUTCUnix", "Score", "NumComments", "IsSelf").
		Updates(&RedditPost{
			Model:          gorm.Model{DeletedAt: gorm.DeletedAt{Time: now, Valid: true}},
			Title:          "",
			Selftext:       "",
			Author:         "",
			Subreddit:      "",
			Permalink:      "",
			URL:            "",
			CreatedUTCUnix: 0,
			Score:          0,
			NumComments:    0,
			IsSelf:         false,
		})
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return ErrRedditPostSoftDeleteNotFound
	}
	return nil
}

func isLikelyUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	s := strings.ToLower(err.Error())
	return strings.Contains(s, "duplicate") ||
		strings.Contains(s, "unique constraint") ||
		strings.Contains(s, "23505")
}

func (r *RedditSource) persistPostIfNew(ctx context.Context, db *gorm.DB, post RedditPostData) (inserted bool, err error) {
	if r.LoadWebsites {
		enqueueURLsFromRedditPost(post)
	}

	pid := strings.TrimSpace(post.ID)

	linked, err := redditSourcePostLinkExists(ctx, db, r.ID, pid)
	if err != nil {
		slog.Error("p_seer_reddit: link lookup", "error", err, "reddit_source_id", r.ID, "post_id", pid)
		return false, err
	}
	if linked {
		return false, nil
	}

	var existing RedditPost
	err = db.WithContext(ctx).Where("post_id = ?", pid).First(&existing).Error
	if err == nil {
		if err := db.Model(r).Association("RedditPosts").Append(&existing); err != nil {
			slog.Error("p_seer_reddit: append existing post to source", "error", err, "reddit_source_id", r.ID, "post_id", pid)
			return false, err
		}
		return true, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		slog.Error("p_seer_reddit: post by post_id lookup", "error", err, "post_id", pid)
		return false, err
	}

	// Soft-deleted row still holds post_id (dedupe tombstone); do not re-create or revive.
	var tomb RedditPost
	if err2 := db.WithContext(ctx).Unscoped().Where("post_id = ?", pid).First(&tomb).Error; err2 == nil && tomb.DeletedAt.Valid {
		return false, nil
	}

	title := strings.TrimSpace(post.Title)
	selftext := strings.TrimSpace(post.Selftext)

	var runnerPtr *uint
	if r.RedditRunnerID != 0 {
		rid := r.RedditRunnerID
		runnerPtr = &rid
	}

	rp := RedditPost{
		RedditRunnerID: runnerPtr,
		PostID:         pid,
		Title:          title,
		Selftext:       selftext,
		Author:         strings.TrimSpace(post.Author),
		Subreddit:      strings.TrimSpace(post.Subreddit),
		Permalink:      strings.TrimSpace(post.Permalink),
		URL:            strings.TrimSpace(post.URL),
		CreatedUTCUnix: post.CreatedUTC,
		Score:          post.Score,
		NumComments:    post.NumComments,
		IsSelf:         post.IsSelf,
	}
	if createErr := db.WithContext(ctx).Create(&rp).Error; createErr != nil {
		if !isLikelyUniqueViolation(createErr) {
			slog.Error("p_seer_reddit: create reddit post", "error", createErr, "reddit_source_id", r.ID)
			return false, createErr
		}
		if reloadErr := db.WithContext(ctx).Unscoped().Where("post_id = ?", pid).First(&existing).Error; reloadErr != nil {
			return false, fmt.Errorf("create reddit post: %w; reload post_id %q: %w", createErr, pid, reloadErr)
		}
		if existing.DeletedAt.Valid {
			return false, nil
		}
		if err := db.Model(r).Association("RedditPosts").Append(&existing); err != nil {
			slog.Error("p_seer_reddit: append after duplicate create", "error", err, "reddit_source_id", r.ID, "post_id", pid)
			return false, err
		}
		return true, nil
	}
	if err := db.Model(r).Association("RedditPosts").Append(&rp); err != nil {
		slog.Error("p_seer_reddit: append new post to source", "error", err, "reddit_source_id", r.ID, "post_id", pid)
		return false, err
	}
	return true, nil
}

// Kind satisfies [github.com/lariv-in/lago/plugins/p_seer_intel.IntelKind] for *RedditPost.
func (p RedditPost) Kind() string {
	return "reddit"
}

// IntelID satisfies [github.com/lariv-in/lago/plugins/p_seer_intel.IntelKind] for *RedditPost.
func (p RedditPost) IntelID() uint {
	return p.ID
}

// Content returns markdown derived from persisted fields (title, selftext, metadata links).
// Satisfies [github.com/lariv-in/lago/plugins/p_seer_intel.IntelKind] for *RedditPost.
func (p *RedditPost) Content() string {
	if p == nil {
		return ""
	}
	var b strings.Builder
	title := strings.TrimSpace(p.Title)
	if title != "" {
		b.WriteString("# ")
		b.WriteString(title)
		b.WriteString("\n\n")
	}
	if txt := strings.TrimSpace(p.Selftext); txt != "" {
		b.WriteString(txt)
		b.WriteString("\n\n")
	}
	b.WriteString("---\n\n")
	fmt.Fprintf(&b, "- **Author:** u/%s\n", p.Author)
	fmt.Fprintf(&b, "- **Subreddit:** r/%s\n", p.Subreddit)
	fmt.Fprintf(&b, "- **Score:** %d\n", p.Score)
	fmt.Fprintf(&b, "- **Comments:** %d\n", p.NumComments)
	if p.Permalink != "" {
		fmt.Fprintf(&b, "- **Permalink:** https://www.reddit.com%s\n", p.Permalink)
	}
	if !p.IsSelf {
		if u := strings.TrimSpace(p.URL); u != "" {
			fmt.Fprintf(&b, "- **Link:** %s\n", u)
		}
	}
	return strings.TrimSpace(b.String())
}

// IntelDetail satisfies [github.com/lariv-in/lago/plugins/p_seer_intel.IntelKind] for *RedditPost: app path to [RedditPostDetailRoute].
func (p *RedditPost) IntelDetail(ctx context.Context) (string, error) {
	if p == nil || p.ID == 0 {
		return "", fmt.Errorf("p_seer_reddit: IntelDetail: missing post")
	}
	return lago.RoutePath("seer_reddit.RedditPostDetailRoute", map[string]getters.Getter[any]{
		"id": getters.Any(getters.Static(strconv.FormatUint(uint64(p.ID), 10))),
	})(ctx)
}
