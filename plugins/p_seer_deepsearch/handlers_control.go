package p_seer_deepsearch

import (
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/views"
	"gorm.io/gorm"
)

func deepSearchDetailRedirectURL(r *http.Request, id uint) (string, error) {
	return lago.RoutePath("seer_deepsearch.DetailRoute", map[string]getters.Getter[any]{
		"id": getters.Any(getters.Static(strconv.FormatUint(uint64(id), 10))),
	})(r.Context())
}

// deepSearchStopPostLayer cancels an in-flight pipeline for this id (see [TryStopDeepSearchPipeline]).
type deepSearchStopPostLayer struct{}

func (deepSearchStopPostLayer) Next(_ views.View, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			next.ServeHTTP(w, r)
			return
		}
		idStr := strings.TrimSpace(r.PathValue("id"))
		id64, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil || id64 == 0 {
			http.Error(w, "invalid id", http.StatusBadRequest)
			return
		}
		id := uint(id64)
		db, dberr := getters.DBFromContext(r.Context())
		if dberr != nil {
			slog.Error("p_seer_deepsearch: stop missing db", "error", dberr)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		if TryStopDeepSearchPipeline(id) {
			appendDeepSearchLog(r.Context(), db, id, DeepSearchLogKindInfo, "stop requested via UI (pipeline will wind down)")
		}
		detailURL, err := deepSearchDetailRedirectURL(r, id)
		if err != nil || detailURL == "" {
			slog.Error("p_seer_deepsearch: stop redirect", "error", err, "deep_search_id", id)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		views.HtmxRedirect(w, r, detailURL, http.StatusSeeOther)
	})
}

// deepSearchRestartPostLayer clears a terminal job and starts [BeginDeepSearchPipeline] again.
type deepSearchRestartPostLayer struct{}

func (deepSearchRestartPostLayer) Next(_ views.View, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			next.ServeHTTP(w, r)
			return
		}
		idStr := strings.TrimSpace(r.PathValue("id"))
		id64, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil || id64 == 0 {
			http.Error(w, "invalid id", http.StatusBadRequest)
			return
		}
		id := uint(id64)
		db, dberr := getters.DBFromContext(r.Context())
		if dberr != nil {
			slog.Error("p_seer_deepsearch: restart missing db", "error", dberr)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		tx := db.WithContext(r.Context()).Model(&DeepSearch{}).Where("id = ? AND status IN ?", id, []string{
			DeepSearchStatusDone, DeepSearchStatusFailed, DeepSearchStatusCancelled,
		}).Updates(map[string]any{
			"status":    DeepSearchStatusPending,
			"report":    "",
			"run_error": "",
		})
		if tx.Error != nil {
			slog.Error("p_seer_deepsearch: restart update", "error", tx.Error, "deep_search_id", id)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		if tx.RowsAffected == 0 {
			http.Error(w, "restart only allowed when search is done, failed, or stopped", http.StatusConflict)
			return
		}
		appendDeepSearchLog(r.Context(), db, id, DeepSearchLogKindInfo, "pipeline restart requested (new run starting)")
		dbCopy := db.Session(&gorm.Session{})
		BeginDeepSearchPipeline(dbCopy, id)
		detailURL, rerr := deepSearchDetailRedirectURL(r, id)
		if rerr != nil || detailURL == "" {
			slog.Error("p_seer_deepsearch: restart redirect", "error", rerr, "deep_search_id", id)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		views.HtmxRedirect(w, r, detailURL, http.StatusSeeOther)
	})
}
