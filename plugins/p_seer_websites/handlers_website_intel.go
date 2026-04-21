package p_seer_websites

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

// websiteAddIntelLayer handles POST to create [p_seer_intel.Intel] from the [Website] in context.
type websiteAddIntelLayer struct{}

func (websiteAddIntelLayer) Next(_ views.View, _ http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		db, dberr := getters.DBFromContext(r.Context())
		if dberr != nil {
			slog.Error("p_seer_websites: add intel missing db", "error", dberr)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		v := r.Context().Value("website")
		site, ok := v.(Website)
		if !ok || site.ID == 0 {
			slog.Error("p_seer_websites: add intel missing website in context")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		kind := (Website{}).Kind()
		exists, err := p_seer_intel.IntelExistsForSource(r.Context(), db, kind, site.ID)
		if err != nil {
			slog.Error("p_seer_websites: add intel exists check", "error", err, "website_id", site.ID)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		detailURL, err := lago.RoutePath("seer_websites.WebsiteDetailRoute", map[string]getters.Getter[any]{
			"id": getters.Any(getters.Static(strconv.FormatUint(uint64(site.ID), 10))),
		})(r.Context())
		if err != nil || detailURL == "" {
			slog.Error("p_seer_websites: add intel redirect URL", "error", err, "website_id", site.ID)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		if exists {
			views.HtmxRedirect(w, r, detailURL, http.StatusSeeOther)
			return
		}
		if websiteIntelIngestActive.CompareAndSwap(false, true) {
			siteCopy := site
			dbCopy := db
			go func() {
				defer websiteIntelIngestActive.Store(false)
				ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
				defer cancel()
				RunWebsiteSingleIntelIngest(ctx, dbCopy, siteCopy)
			}()
		}
		views.HtmxRedirect(w, r, detailURL, http.StatusSeeOther)
	})
}

// websiteAddAllIntelLayer handles POST to enqueue Intel creation for every non-deleted [Website].
type websiteAddAllIntelLayer struct{}

func (websiteAddAllIntelLayer) Next(_ views.View, _ http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		db, dberr := getters.DBFromContext(r.Context())
		if dberr != nil {
			slog.Error("p_seer_websites: add all intel missing db", "error", dberr)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		var sites []Website
		if err := db.WithContext(r.Context()).Where("deleted_at IS NULL").Order("id DESC").Find(&sites).Error; err != nil {
			slog.Error("p_seer_websites: add all intel list", "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		listURL, err := lago.RoutePath("seer_websites.WebsiteListRoute", nil)(r.Context())
		if err != nil || listURL == "" {
			slog.Error("p_seer_websites: add all intel redirect URL", "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		if websiteIntelIngestActive.CompareAndSwap(false, true) {
			sitesCopy := append([]Website(nil), sites...)
			dbCopy := db
			go func() {
				defer websiteIntelIngestActive.Store(false)
				ctx, cancel := context.WithTimeout(context.Background(), 60*time.Minute)
				defer cancel()
				RunWebsitesBulkIntelIngest(ctx, dbCopy, sitesCopy)
			}()
		}
		views.HtmxRedirect(w, r, listURL, http.StatusSeeOther)
	})
}
