package p_lacerate

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html"
	"log/slog"
	"strings"

	"github.com/lariv-in/lago/lago"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// redditListingMaxPages caps extra listing fetches when a page contains duplicates already stored for this source.
const redditListingMaxPages = 25

// intelDedupHashFromRedditPostID builds DedupHash from Reddit listing JSON `id` (short post id).
func intelDedupHashFromRedditPostID(redditID string) string {
	id := strings.TrimSpace(redditID)
	if id == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(id))
	return hex.EncodeToString(sum[:])
}

func lacerateUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "unique constraint") ||
		strings.Contains(msg, "duplicate key value") ||
		strings.Contains(msg, "sqlstate 23505") ||
		strings.Contains(msg, "violates unique constraint")
}

type RedditSource struct {
	gorm.Model
	Subreddits  datatypes.JSON `gorm:"type:json"`
	SearchQuery string
	SourceID    uint   `gorm:"not null;uniqueIndex"`
	Source      Source `gorm:"foreignKey:SourceID"`
}

func redditPreviewImageURL(post RedditPostData) string {
	t := strings.TrimSpace(html.UnescapeString(post.Thumbnail))
	if strings.HasPrefix(t, "http://") || strings.HasPrefix(t, "https://") {
		return t
	}
	return ""
}

func redditPostToMarkdown(ctx context.Context, post RedditPostData) string {
	var b strings.Builder
	title := strings.TrimSpace(post.Title)
	if title != "" {
		b.WriteString("# ")
		b.WriteString(title)
		b.WriteString("\n\n")
	}
	if txt := strings.TrimSpace(post.Selftext); txt != "" {
		b.WriteString(txt)
		b.WriteString("\n\n")
	}
	if !post.IsSelf {
		if u := strings.TrimSpace(post.URL); u != "" {
			if body := fetchPostURLAsMarkdown(ctx, u); body != "" {
				b.WriteString("## Linked article\n\n")
				b.WriteString(body)
				b.WriteString("\n\n")
			}
		}
	}
	b.WriteString("---\n\n")
	fmt.Fprintf(&b, "- **Author:** u/%s\n", post.Author)
	fmt.Fprintf(&b, "- **Subreddit:** r/%s\n", post.Subreddit)
	fmt.Fprintf(&b, "- **Score:** %d (up %d / down %d)\n", post.Score, post.Ups, post.Downs)
	fmt.Fprintf(&b, "- **Comments:** %d\n", post.NumComments)
	if post.Permalink != "" {
		fmt.Fprintf(&b, "- **Permalink:** https://www.reddit.com%s\n", post.Permalink)
	}
	if !post.IsSelf {
		if u := strings.TrimSpace(post.URL); u != "" {
			fmt.Fprintf(&b, "- **Link:** %s\n", u)
		}
	}
	return strings.TrimSpace(b.String())
}

func (r RedditSource) fetchSubredditListings(subredditName, searchQuery string, ctx context.Context, db *gorm.DB, out *[]Intel) error {
	afterStr := ""
	for range redditListingMaxPages {
		var afterPtr *string
		if afterStr != "" {
			afterPtr = &afterStr
		}
		var listing *RedditObject[RedditListing[RedditPostData]]
		var err error
		if searchQuery != "" {
			listing, err = FetchSubredditPostsSearch(subredditName, searchQuery, afterPtr)
		} else {
			listing, err = FetchSubredditPosts(subredditName, afterPtr)
		}
		if err != nil {
			return err
		}
		pageHadDup := false
		for _, child := range listing.Data.Children {
			post := child.Data
			dedupe := intelDedupHashFromRedditPostID(post.ID)
			if dedupe == "" {
				slog.Warn("lacerate: reddit post missing id, skip", "subreddit", subredditName)
				continue
			}
			var n int64
			if err := db.Model(&Intel{}).Where("source_id = ? AND dedup_hash = ?", r.SourceID, dedupe).Count(&n).Error; err != nil {
				slog.Error("lacerate: reddit source dedupe count", "error", err, "source_id", r.SourceID)
				return err
			}
			if n > 0 {
				pageHadDup = true
				continue
			}
			previewURL := redditPreviewImageURL(post)
			var previewID *uint
			if previewURL != "" {
				previewID = persistRedditPreviewImage(ctx, db, post, previewURL)
			}
			dedupeCopy := dedupe
			i := Intel{
				SourceID:       r.SourceID,
				DedupHash:      &dedupeCopy,
				Content:        redditPostToMarkdown(ctx, post),
				PreviewImageID: previewID,
			}
			if err := db.Create(&i).Error; err != nil {
				if lacerateUniqueViolation(err) {
					pageHadDup = true
					continue
				}
				slog.Error("lacerate: reddit source create intel", "error", err, "source_id", r.SourceID)
				return err
			}
			*out = append(*out, i)
		}
		if listing.Data.After == nil {
			break
		}
		if !pageHadDup {
			break
		}
		afterStr = *listing.Data.After
	}
	return nil
}

func (r RedditSource) Fetch(ctx context.Context, db *gorm.DB) ([]Intel, error) {
	var subs []string
	if len(r.Subreddits) > 0 {
		if err := json.Unmarshal(r.Subreddits, &subs); err != nil {
			slog.Error("lacerate: reddit source unmarshal subreddits", "error", err, "source_id", r.SourceID)
			return nil, err
		}
	}
	query := strings.TrimSpace(r.SearchQuery)
	var out []Intel
	for _, raw := range subs {
		name := strings.TrimSpace(strings.TrimPrefix(raw, "r/"))
		if name == "" {
			continue
		}
		if err := r.fetchSubredditListings(name, query, ctx, db, &out); err != nil {
			return nil, err
		}
	}
	return out, nil
}

func init() {
	SourceKindMap["reddit"] = SourceDesc{
		Name:  "Reddit",
		Model: RedditSource{},
	}
}

func init() {
	lago.OnDBInit("p_lacerate.reddit_source_model", func(db *gorm.DB) *gorm.DB {
		lago.RegisterModel[RedditSource](db)
		return db
	})
}
