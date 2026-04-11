package p_lacerate

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/views"
	"gorm.io/gorm"
)

// lookupWorkerPOSTOnlyLayer rejects non-POST for lookup worker action routes.
type lookupWorkerPOSTOnlyLayer struct{}

func (lookupWorkerPOSTOnlyLayer) Next(view views.View, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func lookupRestartWorkerHandler(v *views.View) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		idStr := r.PathValue("id")
		id64, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		rawDB := r.Context().Value("$db")
		db, ok := rawDB.(*gorm.DB)
		if !ok || db == nil {
			slog.Error("lacerate: lookup restart worker missing db", "lookup_id", idStr)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		RunLookupNow(db, uint(id64))
		detailURL, err := lago.RoutePath("lacerate.LookupDetailRoute", map[string]getters.Getter[any]{
			"id": getters.Any(getters.Static(idStr)),
		})(r.Context())
		if err != nil || detailURL == "" {
			slog.Error("lacerate: lookup restart worker detail URL", "error", err, "lookup_id", idStr)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		views.HtmxRedirect(w, r, detailURL, http.StatusSeeOther)
	})
}

func lookupStopWorkerHandler(v *views.View) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		idStr := r.PathValue("id")
		id64, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		StopLookupWorker(uint(id64))
		detailURL, err := lago.RoutePath("lacerate.LookupDetailRoute", map[string]getters.Getter[any]{
			"id": getters.Any(getters.Static(idStr)),
		})(r.Context())
		if err != nil || detailURL == "" {
			slog.Error("lacerate: lookup stop worker detail URL", "error", err, "lookup_id", idStr)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		views.HtmxRedirect(w, r, detailURL, http.StatusSeeOther)
	})
}

func init() {
	patchers := views.QueryPatchers[Lookup]{
		{Key: "lacerate.lookups.order_id", Value: views.QueryPatcherOrderBy[Lookup]{Order: "id DESC"}},
	}

	lago.RegistryView.Register("lacerate.LookupListView",
		lago.GetPageView("lacerate.LookupsTable").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("lacerate.lookups.list", views.LayerList[Lookup]{
				Key:           getters.Static("lookups"),
				QueryPatchers: patchers,
			}))

	lago.RegistryView.Register("lacerate.LookupDetailView",
		lago.GetPageView("lacerate.LookupDetail").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("lacerate.lookups.detail", views.LayerDetail[Lookup]{
				Key:          getters.Static("lookup"),
				PathParamKey: getters.Static("id"),
			}).
			WithLayer("lacerate.lookups.detail_logs", lookupDetailLogsLayer{}))

	lago.RegistryView.Register("lacerate.LookupRestartWorkerView",
		lago.GetPageView("lacerate.LookupDetail").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("lacerate.lookups.restart_worker_post_only", lookupWorkerPOSTOnlyLayer{}).
			WithLayer("lacerate.lookups.restart_worker", views.MethodLayer{
				Method:  http.MethodPost,
				Handler: lookupRestartWorkerHandler,
			}))

	lago.RegistryView.Register("lacerate.LookupStopWorkerView",
		lago.GetPageView("lacerate.LookupDetail").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("lacerate.lookups.stop_worker_post_only", lookupWorkerPOSTOnlyLayer{}).
			WithLayer("lacerate.lookups.stop_worker", views.MethodLayer{
				Method:  http.MethodPost,
				Handler: lookupStopWorkerHandler,
			}))

	lago.RegistryView.Register("lacerate.LookupCreateView",
		lago.GetPageView("lacerate.LookupCreateForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("lacerate.lookups.create", views.LayerCreate[Lookup]{
				SuccessURL: lago.RoutePath("lacerate.LookupDetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("$id")),
				}),
			}))

	lago.RegistryView.Register("lacerate.LookupUpdateView",
		lago.GetPageView("lacerate.LookupUpdateForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("lacerate.lookups.update_detail", views.LayerDetail[Lookup]{
				Key:          getters.Static("lookup"),
				PathParamKey: getters.Static("id"),
			}).
			WithLayer("lacerate.lookups.update", views.LayerUpdate[Lookup]{
				Key: getters.Static("lookup"),
				SuccessURL: lago.RoutePath("lacerate.LookupDetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("lookup.ID")),
				}),
			}))

	lago.RegistryView.Register("lacerate.LookupDeleteView",
		lago.GetPageView("lacerate.LookupDeleteForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("lacerate.lookups.delete_detail", views.LayerDetail[Lookup]{
				Key:          getters.Static("lookup"),
				PathParamKey: getters.Static("id"),
			}).
			WithLayer("lacerate.lookups.delete", views.LayerDelete[Lookup]{
				Key:        getters.Static("lookup"),
				SuccessURL: lago.RoutePath("lacerate.LookupListRoute", nil),
			}))
}
