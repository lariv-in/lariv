package p_seer_websites

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/views"
)

type websiteRunnerWorkerPoolPOSTOnlyLayer struct{}

func (websiteRunnerWorkerPoolPOSTOnlyLayer) Next(view views.View, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func websiteRunnerWorkerPoolStartHandler(_ *views.View) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		idStr := r.PathValue("id")
		id64, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil || id64 == 0 {
			http.NotFound(w, r)
			return
		}
		runnerID := uint(id64)
		db, dberr := getters.DBFromContext(r.Context())
		if dberr != nil {
			slog.Error("p_seer_websites: worker pool start missing db", "error", dberr, "runner_id", runnerID)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		ScheduleWebsiteRunnerWorkerPoolStart(db, runnerID)
		detailURL, err := lago.RoutePath("seer_websites.WebsiteRunnerDetailRoute", map[string]getters.Getter[any]{
			"id": getters.Any(getters.Static(idStr)),
		})(r.Context())
		if err != nil || detailURL == "" {
			slog.Error("p_seer_websites: worker pool start detail URL", "error", err, "runner_id", runnerID)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		views.HtmxRedirect(w, r, detailURL, http.StatusSeeOther)
	})
}

func websiteRunnerWorkerPoolStopHandler(_ *views.View) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		idStr := r.PathValue("id")
		id64, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil || id64 == 0 {
			http.NotFound(w, r)
			return
		}
		StopWebsiteRunnerWorkerPool(uint(id64))
		detailURL, err := lago.RoutePath("seer_websites.WebsiteRunnerDetailRoute", map[string]getters.Getter[any]{
			"id": getters.Any(getters.Static(idStr)),
		})(r.Context())
		if err != nil || detailURL == "" {
			slog.Error("p_seer_websites: worker pool stop detail URL", "error", err, "runner_id", idStr)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		views.HtmxRedirect(w, r, detailURL, http.StatusSeeOther)
	})
}

func registerWebsiteRunnerWorkerPoolViews() {
	lago.RegistryView.Register("seer_websites.WebsiteRunnerWorkerPoolStartView",
		lago.GetPageView("seer_websites.WebsiteRunnerDetail").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_websites.website_runner.worker_pool_start_post", websiteRunnerWorkerPoolPOSTOnlyLayer{}).
			WithLayer("seer_websites.website_runner.worker_pool_start", views.MethodLayer{
				Method:  http.MethodPost,
				Handler: websiteRunnerWorkerPoolStartHandler,
			}))

	lago.RegistryView.Register("seer_websites.WebsiteRunnerWorkerPoolStopView",
		lago.GetPageView("seer_websites.WebsiteRunnerDetail").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_websites.website_runner.worker_pool_stop_post", websiteRunnerWorkerPoolPOSTOnlyLayer{}).
			WithLayer("seer_websites.website_runner.worker_pool_stop", views.MethodLayer{
				Method:  http.MethodPost,
				Handler: websiteRunnerWorkerPoolStopHandler,
			}))
}
