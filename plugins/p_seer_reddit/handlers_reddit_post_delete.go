package p_seer_reddit

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/views"
)

// redditPostSoftDeleteLayer handles POST to wipe and soft-delete a [RedditPost] by path id.
type redditPostSoftDeleteLayer struct{}

func (redditPostSoftDeleteLayer) Next(_ views.View, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			next.ServeHTTP(w, r)
			return
		}
		db, dberr := getters.DBFromContext(r.Context())
		if dberr != nil {
			slog.Error("p_seer_reddit: soft delete missing db", "error", dberr)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		raw := r.PathValue("id")
		id64, err := strconv.ParseUint(raw, 10, 64)
		if err != nil || id64 == 0 {
			http.NotFound(w, r)
			return
		}
		id := uint(id64)
		if err := SoftWipeRedditPost(r.Context(), db, id); err != nil {
			if errors.Is(err, ErrRedditPostSoftDeleteNotFound) {
				http.NotFound(w, r)
				return
			}
			slog.Error("p_seer_reddit: soft delete post", "error", err, "post_id", id)
			http.Error(w, "Delete failed", http.StatusBadGateway)
			return
		}
		listURL, err := lago.RoutePath("seer_reddit.RedditPostListRoute", nil)(r.Context())
		if err != nil || listURL == "" {
			slog.Error("p_seer_reddit: soft delete redirect URL", "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		views.HtmxRedirect(w, r, listURL, http.StatusSeeOther)
	})
}
