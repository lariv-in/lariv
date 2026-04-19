package p_seer_deepsearch

import (
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/views"
)

// deepSearchStartPostLayer creates a [DeepSearch] row and runs [runDeepSearchPipeline] in a background goroutine.
type deepSearchStartPostLayer struct{}

func (deepSearchStartPostLayer) Next(_ views.View, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			next.ServeHTTP(w, r)
			return
		}
		if err := r.ParseForm(); err != nil {
			slog.Error("p_seer_deepsearch: parse form", "error", err)
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}
		q := strings.TrimSpace(r.PostFormValue("Query"))
		if q == "" {
			http.Error(w, "Query is required", http.StatusBadRequest)
			return
		}
		db, dberr := getters.DBFromContext(r.Context())
		if dberr != nil {
			slog.Error("p_seer_deepsearch: start missing db", "error", dberr)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		row := DeepSearch{
			Query:  q,
			Status: DeepSearchStatusPending,
		}
		if err := db.WithContext(r.Context()).Create(&row).Error; err != nil {
			slog.Error("p_seer_deepsearch: create row", "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		id := row.ID
		dbCopy := db
		BeginDeepSearchPipeline(dbCopy, id)

		detailURL, err := lago.RoutePath("seer_deepsearch.DetailRoute", map[string]getters.Getter[any]{
			"id": getters.Any(getters.Static(strconv.FormatUint(uint64(id), 10))),
		})(r.Context())
		if err != nil || detailURL == "" {
			slog.Error("p_seer_deepsearch: detail redirect URL", "error", err, "deep_search_id", id)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		views.HtmxRedirect(w, r, detailURL, http.StatusSeeOther)
	})
}

// deepSearchStartRejectGetLayer responds with 405 for non-POST on the start route.
type deepSearchStartRejectGetLayer struct{}

func (deepSearchStartRejectGetLayer) Next(_ views.View, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		next.ServeHTTP(w, r)
	})
}
