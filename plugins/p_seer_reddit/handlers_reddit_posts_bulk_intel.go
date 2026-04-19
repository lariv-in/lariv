package p_seer_reddit

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/views"
)

// redditPostsBulkAddIntelLayer runs after [views.LayerList] has set `redditPosts` in context.
// POST starts async bulk ingest; GET returns 405.
type redditPostsBulkAddIntelLayer struct {
	redirectRouteName string
	sourceIDPathParam string
}

func (l redditPostsBulkAddIntelLayer) Next(_ views.View, _ http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		db, dberr := getters.DBFromContext(r.Context())
		if dberr != nil {
			slog.Error("p_seer_reddit: bulk add intel missing db", "error", dberr)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		v := r.Context().Value("redditPosts")
		ol, ok := v.(components.ObjectList[RedditPost])
		if !ok {
			slog.Error("p_seer_reddit: bulk add intel missing redditPosts in context")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		if p := l.sourceIDPathParam; p != "" {
			raw := r.PathValue(p)
			id64, err := strconv.ParseUint(raw, 10, 64)
			if err != nil || id64 == 0 {
				http.NotFound(w, r)
				return
			}
		}
		if redditIntelIngestActive.CompareAndSwap(false, true) {
			items := append([]RedditPost(nil), ol.Items...)
			dbCopy := db
			go func() {
				defer redditIntelIngestActive.Store(false)
				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
				defer cancel()
				RunRedditBulkIntelIngest(ctx, dbCopy, items)
			}()
		}
		var redirectArgs map[string]getters.Getter[any]
		if p := l.sourceIDPathParam; p != "" {
			redirectArgs = map[string]getters.Getter[any]{
				"source_id": getters.Any(getters.Static(r.PathValue(p))),
			}
		}
		baseURL, err := lago.RoutePath(l.redirectRouteName, redirectArgs)(r.Context())
		if err != nil || baseURL == "" {
			slog.Error("p_seer_reddit: bulk add intel redirect URL", "error", err, "route", l.redirectRouteName)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		if q := r.URL.RawQuery; q != "" {
			baseURL += "?" + q
		}
		views.HtmxRedirect(w, r, baseURL, http.StatusSeeOther)
	})
}
