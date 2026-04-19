package p_seer_reddit

import (
	"net/http"

	"github.com/lariv-in/lago/views"
	"gorm.io/gorm"
)

// redditPostsForCurrentSourcePatcher scopes [RedditPost] queries to posts linked to the [RedditSource]
// in context (set by [views.LayerDetail] on by-source routes).
type redditPostsForCurrentSourcePatcher struct{}

func (redditPostsForCurrentSourcePatcher) Patch(_ views.View, r *http.Request, q gorm.ChainInterface[RedditPost]) gorm.ChainInterface[RedditPost] {
	var sid uint
	switch v := r.Context().Value("redditSource").(type) {
	case RedditSource:
		sid = v.ID
	case *RedditSource:
		if v != nil {
			sid = v.ID
		}
	}
	if sid == 0 {
		// No source in context or ID not loaded — empty set only (unscoped query would be every post).
		return q.Where("1 = 0")
	}
	return q.Where(
		"id IN (SELECT reddit_post_id FROM "+RedditSourcePostsJoinTable+" WHERE reddit_source_id = ?)",
		sid,
	)
}
