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

// redditListingMaxPages caps Reddit listing pagination per subreddit (safety bound).
const redditListingMaxPages = 25

type RedditSource struct {
	gorm.Model
	Subreddits     datatypes.JSON `gorm:"type:json"`
	SearchQuery    string
	MaxFreshPosts  uint `gorm:"not null;default:25"`
	SourceID       uint `gorm:"not null;uniqueIndex"`
	Source         Source `gorm:"foreignKey:SourceID"`
}

func (r RedditSource) effectiveMaxFreshPosts() int {
	return sourceEffectiveMaxFreshPosts(r.MaxFreshPosts)
}

func (r RedditSource) fetchSubredditListings(subredditName, searchQuery string, ctx context.Context, db *gorm.DB, existingDedup map[string]struct{}, out *[]Intel, freshTotal *int, maxFresh int) error {
	afterStr := ""
	for range redditListingMaxPages {
		if *freshTotal >= maxFresh {
			return nil
		}
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
		for _, child := range listing.Data.Children {
			if *freshTotal >= maxFresh {
				return nil
			}
			post := child.Data
			dedupe := post.IntelDedupHash()
			if dedupe == "" {
				slog.Warn("lacerate: reddit post missing id, skip", "subreddit", subredditName)
				continue
			}
			if _, dup := existingDedup[dedupe]; dup {
				continue
			}
			previewURL := post.PreviewImageURL()
			var previewID *uint
			if previewURL != "" {
				previewID = persistRedditPreviewImage(ctx, db, post, previewURL)
			}
			dedupeCopy := dedupe
			sourceID := r.SourceID
			i := Intel{
				SourceID:       &sourceID,
				DedupHash:      &dedupeCopy,
				Content:        post.Markdown(ctx),
				PreviewImageID: previewID,
			}
			existingDedup[dedupe] = struct{}{}
			*out = append(*out, i)
			*freshTotal++
		}
		if listing.Data.After == nil {
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
	maxFresh := r.effectiveMaxFreshPosts()
	var out []Intel
	freshTotal := 0
	for _, raw := range subs {
		if freshTotal >= maxFresh {
			break
		}
		name := strings.TrimSpace(strings.TrimPrefix(raw, "r/"))
		if name == "" {
			continue
		}
		if err := r.fetchSubredditListings(name, query, ctx, db, existingDedup, &out, &freshTotal, maxFresh); err != nil {
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
