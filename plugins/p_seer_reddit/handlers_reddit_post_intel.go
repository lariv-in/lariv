package p_seer_reddit

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_seer_intel"
	"github.com/lariv-in/lago/views"
)

// redditPostAddIntelLayer handles POST to create an [p_seer_intel.Intel] from the [RedditPost] in context.
// Non-POST returns 405. If intel already exists, redirects without starting a job. Otherwise starts async ingest.
type redditPostAddIntelLayer struct{}

func (redditPostAddIntelLayer) Next(_ views.View, _ http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		db, dberr := getters.DBFromContext(r.Context())
		if dberr != nil {
			slog.Error("p_seer_reddit: add intel missing db", "error", dberr)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		v := r.Context().Value("redditPost")
		post, ok := v.(RedditPost)
		if !ok || post.ID == 0 {
			slog.Error("p_seer_reddit: add intel missing redditPost in context")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		kind := (RedditPost{}).Kind()
		exists, err := p_seer_intel.IntelExistsForSource(r.Context(), db, kind, post.ID)
		if err != nil {
			slog.Error("p_seer_reddit: add intel exists check", "error", err, "post_id", post.ID)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		detailURL, err := lago.RoutePath("seer_reddit.RedditPostDetailRoute", map[string]getters.Getter[any]{
			"id": getters.Any(getters.Static(strconv.FormatUint(uint64(post.ID), 10))),
		})(r.Context())
		if err != nil || detailURL == "" {
			slog.Error("p_seer_reddit: add intel redirect URL", "error", err, "post_id", post.ID)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		if exists {
			views.HtmxRedirect(w, r, detailURL, http.StatusSeeOther)
			return
		}
		if redditIntelIngestActive.CompareAndSwap(false, true) {
			postCopy := post
			dbCopy := db
			go func() {
				defer redditIntelIngestActive.Store(false)
				ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
				defer cancel()
				RunRedditSinglePostIntelIngest(ctx, dbCopy, postCopy)
			}()
		}
		views.HtmxRedirect(w, r, detailURL, http.StatusSeeOther)
	})
}
