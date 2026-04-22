package p_seer_websites

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

type websiteSourceFetchActionLayer struct{}

func (websiteSourceFetchActionLayer) Next(_ views.View, _ http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		db, dberr := getters.DBFromContext(r.Context())
		if dberr != nil {
			slog.Error("p_seer_websites: fetch source missing db", "error", dberr)
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
		var src WebsiteSource
		if err := db.WithContext(r.Context()).First(&src, id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				http.NotFound(w, r)
				return
			}
			slog.Error("p_seer_websites: fetch source load", "error", err, "website_source_id", id)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		if WebsiteSourceCrawlIsRunning(id) {
			slog.Info("p_seer_websites: fetch source skipped (crawl already running)", "website_source_id", id)
			backURL, rerr := lago.RoutePath("seer_websites.WebsiteSourceDetailRoute", map[string]getters.Getter[any]{
				"id": getters.Any(getters.Static(raw)),
			})(r.Context())
			if rerr != nil || backURL == "" {
				slog.Error("p_seer_websites: fetch source busy redirect", "error", rerr, "website_source_id", id)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
			views.HtmxRedirect(w, r, backURL, http.StatusSeeOther)
			return
		}
		srcCopy := src
		dbCopy := db
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Minute)
			defer cancel()
			if err := srcCopy.Fetch(ctx, dbCopy); err != nil {
				slog.Error("p_seer_websites: fetch source async", "error", err, "website_source_id", srcCopy.ID)
			}
		}()

		backURL, err := lago.RoutePath("seer_websites.WebsiteSourceDetailRoute", map[string]getters.Getter[any]{
			"id": getters.Any(getters.Static(raw)),
		})(r.Context())
		if err != nil || backURL == "" {
			slog.Error("p_seer_websites: fetch source redirect", "error", err, "website_source_id", id)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		views.HtmxRedirect(w, r, backURL, http.StatusSeeOther)
	})
}
