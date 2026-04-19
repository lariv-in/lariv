package p_seer_reddit

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/views"
	"gorm.io/gorm"
)

// redditSourceFetchPostsActionLayer handles POST to fetch new Reddit posts for the source in the path (non-POST → 405).
type redditSourceFetchPostsActionLayer struct{}

func (redditSourceFetchPostsActionLayer) Next(_ views.View, _ http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		db, dberr := getters.DBFromContext(r.Context())
		if dberr != nil {
			slog.Error("p_seer_reddit: fetch posts missing db", "error", dberr)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		raw := r.PathValue("source_id")
		id64, err := strconv.ParseUint(raw, 10, 64)
		if err != nil || id64 == 0 {
			http.NotFound(w, r)
			return
		}
		id := uint(id64)
		var src RedditSource
		if err := db.WithContext(r.Context()).First(&src, id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				http.NotFound(w, r)
				return
			}
			slog.Error("p_seer_reddit: fetch posts load source", "error", err, "source_id", id)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		if redditFetchPostsActive.CompareAndSwap(false, true) {
			srcCopy := src
			dbCopy := db
			go func() {
				defer redditFetchPostsActive.Store(false)
				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
				defer cancel()
				if err := FetchNewRedditPosts(ctx, dbCopy, &srcCopy); err != nil {
					slog.Error("p_seer_reddit: fetch posts async", "error", err, "source_id", srcCopy.ID)
				}
			}()
		}
		backURL, err := lago.RoutePath("seer_reddit.RedditPostListBySourceRoute", map[string]getters.Getter[any]{
			"source_id": getters.Any(getters.Static(raw)),
		})(r.Context())
		if err != nil || backURL == "" {
			slog.Error("p_seer_reddit: fetch posts redirect URL", "error", err, "source_id", id)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		views.HtmxRedirect(w, r, backURL, http.StatusSeeOther)
	})
}
