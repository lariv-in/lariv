package p_seer_reddit

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/lariv-in/lago/plugins/p_seer_intel"
	"gorm.io/gorm"
)

const defaultMaxFreshPosts = 25

func effectiveMaxFreshPosts(stored uint) int {
	n := stored
	if n == 0 {
		n = defaultMaxFreshPosts
	}
	return int(n)
}

// FetchNewRedditPosts loads Reddit JSON listings and inserts new [RedditPost] rows only.
// Intel rows are not created here; a separate pipeline may create [p_seer_intel.Intel] and set [RedditPost.IntelID] later.
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
	maxFresh := effectiveMaxFreshPosts(src.MaxFreshPosts)
	freshTotal := 0

	for _, raw := range subs {
		if freshTotal >= maxFresh {
			break
		}
		name := strings.TrimSpace(strings.TrimPrefix(raw, "r/"))
		if name == "" {
			continue
		}
		if err := src.fetchSubredditListings(ctx, db, name, query, maxFresh, &freshTotal); err != nil {
			return err
		}
	}
	return nil
}

func (r *RedditSource) fetchSubredditListings(ctx context.Context, db *gorm.DB, subredditName, searchQuery string, maxFresh int, freshTotal *int) error {
	afterStr := ""
	for range redditListingMaxPages {
		if *freshTotal >= maxFresh {
			return nil
		}
		var afterPtr *string
		if afterStr != "" {
			afterPtr = &afterStr
		}
		var listing *redditObject[redditListing[RedditPostData]]
		var err error
		if searchQuery != "" {
			listing, err = fetchSubredditPostsSearch(subredditName, searchQuery, afterPtr)
		} else {
			listing, err = fetchSubredditPosts(subredditName, afterPtr)
		}
		if err != nil {
			return err
		}
		for _, child := range listing.Data.Children {
			if *freshTotal >= maxFresh {
				return nil
			}
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
			inserted, err := r.persistPostIfNew(ctx, db, post)
			if err != nil {
				return err
			}
			if inserted {
				*freshTotal++
			}
		}
		if listing.Data.After == nil {
			break
		}
		afterStr = *listing.Data.After
	}
	return nil
}

func (r *RedditSource) persistPostIfNew(ctx context.Context, db *gorm.DB, post RedditPostData) (inserted bool, err error) {
	pid := strings.TrimSpace(post.ID)
	var existing RedditPost
	err = db.WithContext(ctx).Where("reddit_source_id = ? AND post_id = ?", r.ID, pid).First(&existing).Error
	if err == nil {
		return false, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		slog.Error("p_seer_reddit: dedupe lookup", "error", err, "reddit_source_id", r.ID, "post_id", pid)
		return false, err
	}

	title := strings.TrimSpace(post.Title)
	selftext := strings.TrimSpace(post.Selftext)

	rp := RedditPost{
		RedditSourceID: r.ID,
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
	if err := db.WithContext(ctx).Create(&rp).Error; err != nil {
		slog.Error("p_seer_reddit: create reddit post", "error", err, "reddit_source_id", r.ID)
		return false, err
	}
	return true, nil
}

// Content implements [p_seer_intel.IntelKind] using persisted post fields (no live HTTP).
func (p *RedditPost) Content() string {
	return redditPostMarkdownFromStored(p)
}

func redditPostMarkdownFromStored(p *RedditPost) string {
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

var _ p_seer_intel.IntelKind = (*RedditPost)(nil)
