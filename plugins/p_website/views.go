package p_website

import (
	"errors"
	"net/http"

	"github.com/lariv-in/lariv"
	"github.com/lariv-in/lariv/getters"
	"github.com/lariv-in/lariv/plugins/p_users"
	"github.com/lariv-in/lariv/registry"
	"github.com/lariv-in/lariv/views"
	"gorm.io/gorm"
)

type DynamicRouteLayer struct{}

func (m DynamicRouteLayer) Next(view views.View, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		db, err := getters.DBFromContext(r.Context())
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		_, err = FindMatchingDBRoute(db, r.URL.Path)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				http.NotFound(w, r)
				return
			}
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func pluginViews() lariv.PluginFeatures[*views.View] {
	return lariv.PluginFeatures[*views.View]{
		Entries: []registry.Pair[string, *views.View]{
			{
				Key: "p_website.DynamicWebsiteView",
				Value: lariv.GetPageView("p_website.DynamicWebsitePage").
					WithLayer("p_website.DynamicRouteLayer", DynamicRouteLayer{}),
			},
			{
				Key: "p_website.RoutesListView",
				Value: lariv.GetPageView("p_website.RoutesListPage").
					WithLayer("p_users.auth", p_users.AuthenticationLayer{}).
					WithLayer("p_website.routes.list", views.LayerList[DBRoute]{
						Key: getters.Static("dbroutes"),
						QueryPatchers: views.QueryPatchers[DBRoute]{
							{Key: "p_website.routes.preload", Value: views.QueryPatcherPreload[DBRoute]{Fields: []string{"Page", "References"}}},
						},
					}),
			},
			{
				Key: "p_website.RoutesCreateView",
				Value: lariv.GetPageView("p_website.RoutesCreatePage").
					WithLayer("p_users.auth", p_users.AuthenticationLayer{}).
					WithLayer("p_website.routes.create", views.LayerCreate[DBRoute]{
						SuccessURL: lariv.RoutePath("p_website.RoutesListRoute", nil),
					}),
			},
			{
				Key: "p_website.RoutesDetailView",
				Value: lariv.GetPageView("p_website.RoutesDetailPage").
					WithLayer("p_users.auth", p_users.AuthenticationLayer{}).
					WithLayer("p_website.routes.detail", views.LayerDetail[DBRoute]{
						Key:          getters.Static("dbroute"),
						PathParamKey: getters.Static("id"),
						QueryPatchers: views.QueryPatchers[DBRoute]{
							{Key: "p_website.routes.preload", Value: views.QueryPatcherPreload[DBRoute]{Fields: []string{"Page", "References"}}},
						},
					}),
			},
			{
				Key: "p_website.RoutesUpdateView",
				Value: lariv.GetPageView("p_website.RoutesUpdatePage").
					WithLayer("p_users.auth", p_users.AuthenticationLayer{}).
					WithLayer("p_website.routes.detail", views.LayerDetail[DBRoute]{
						Key:          getters.Static("dbroute"),
						PathParamKey: getters.Static("id"),
						QueryPatchers: views.QueryPatchers[DBRoute]{
							{Key: "p_website.routes.preload", Value: views.QueryPatcherPreload[DBRoute]{Fields: []string{"Page", "References"}}},
						},
					}).
					WithLayer("p_website.routes.update", views.LayerUpdate[DBRoute]{
						Key: getters.Static("dbroute"),
						SuccessURL: lariv.RoutePath("p_website.RoutesDetailRoute", map[string]getters.Getter[any]{
							"id": getters.Any(getters.Key[uint]("dbroute.ID")),
						}),
					}),
			},
			{
				Key: "p_website.RoutesDeleteView",
				Value: lariv.GetPageView("p_website.RoutesDeleteForm").
					WithLayer("p_users.auth", p_users.AuthenticationLayer{}).
					WithLayer("p_website.routes.detail", views.LayerDetail[DBRoute]{
						Key:          getters.Static("dbroute"),
						PathParamKey: getters.Static("id"),
					}).
					WithLayer("p_website.routes.delete", views.LayerDelete[DBRoute]{
						Key:        getters.Static("dbroute"),
						SuccessURL: lariv.RoutePath("p_website.RoutesListRoute", nil),
					}),
			},
		},
	}
}
