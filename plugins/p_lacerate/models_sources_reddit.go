package p_lacerate

import (
	"context"
	"encoding/json"
	"log/slog"
	"strings"

	"github.com/lariv-in/lago/lago"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// redditListingMaxPages caps extra listing fetches when a page contains duplicates already known for this source.
const redditListingMaxPages = 25

type RedditSource struct {
	gorm.Model
	Subreddits  datatypes.JSON `gorm:"type:json"`
	SearchQuery string
	SourceID    uint   `gorm:"not null;uniqueIndex"`
	Source      Source `gorm:"foreignKey:SourceID"`
}

func (r RedditSource) fetchSubredditListings(subredditName, searchQuery string, ctx context.Context, db *gorm.DB, existingDedup map[string]struct{}, out *[]Intel) error {
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
			dedupe := post.IntelDedupHash()
			if dedupe == "" {
				slog.Warn("lacerate: reddit post missing id, skip", "subreddit", subredditName)
				continue
			}
			if _, dup := existingDedup[dedupe]; dup {
				pageHadDup = true
				continue
			}
			previewURL := post.PreviewImageURL()
			var previewID *uint
			if previewURL != "" {
				previewID = persistRedditPreviewImage(ctx, db, post, previewURL)
			}
			dedupeCopy := dedupe
			i := Intel{
				SourceID:       r.SourceID,
				DedupHash:      &dedupeCopy,
				Content:        post.Markdown(ctx),
				PreviewImageID: previewID,
			}
			existingDedup[dedupe] = struct{}{}
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

func (r RedditSource) Fetch(ctx context.Context, db *gorm.DB, existingDedup map[string]struct{}) ([]Intel, error) {
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
		if err := r.fetchSubredditListings(name, query, ctx, db, existingDedup, &out); err != nil {
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
	if err := RegistrySourceKind.Register("reddit", func() SourceInterface { return &RedditSource{} }); err != nil {
		panic(err)
	}
}

func init() {
	lago.OnDBInit("p_lacerate.reddit_source_model", func(db *gorm.DB) *gorm.DB {
		lago.RegisterModel[RedditSource](db)
		return db
	})
}
