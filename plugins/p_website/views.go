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

func (m DynamicRouteLayer) Next(_ *views.View, next http.Handler) http.Handler {
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
						FormPatchers: views.FormPatchers{
							{Key: "p_website.create_blank_page", Value: createBlankPagePatcher{}},
						},
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
			{
				Key: "p_website.RoutesBuilderView",
				Value: lariv.GetPageView("p_website.RoutesBuilderPage").
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
				Key: "p_website.RoutesBuilderProjectView",
				Value: lariv.GetPageView("p_website.RoutesBuilderPage").
					WithLayer("p_users.auth", p_users.AuthenticationLayer{}).
					WithLayer("p_website.routes.detail", views.LayerDetail[DBRoute]{
						Key:          getters.Static("dbroute"),
						PathParamKey: getters.Static("id"),
						QueryPatchers: views.QueryPatchers[DBRoute]{
							{Key: "p_website.routes.preload", Value: views.QueryPatcherPreload[DBRoute]{Fields: []string{"Page", "References"}}},
						},
					}).
					WithLayer("p_website.builder.load", views.MethodLayer{
						Method:  http.MethodGet,
						Handler: loadBuilderProjectHandler,
					}).
					WithLayer("p_website.builder.store", views.MethodLayer{
						Method:  http.MethodPost,
						Handler: storeBuilderProjectHandler,
					}),
			},
			{
				Key: "p_website.RoutesBuilderThemeView",
				Value: lariv.GetPageView("p_website.RoutesBuilderPage").
					WithLayer("p_users.auth", p_users.AuthenticationLayer{}).
					WithLayer("p_website.routes.detail", views.LayerDetail[DBRoute]{
						Key:          getters.Static("dbroute"),
						PathParamKey: getters.Static("id"),
					}).
					WithLayer("p_website.builder.theme", views.MethodLayer{
						Method:  http.MethodPost,
						Handler: storeBuilderThemeHandler,
					}),
			},
			{
				Key: "p_website.BuilderAssetUploadView",
				Value: lariv.GetPageView("p_website.RoutesBuilderPage").
					WithLayer("p_users.auth", p_users.AuthenticationLayer{}).
					WithLayer("p_website.builder.asset_upload", views.MethodLayer{
						Method:  http.MethodPost,
						Handler: builderAssetUploadHandler,
					}),
			},
			{
				Key: "p_website.PublicAssetView",
				Value: lariv.GetPageView("p_website.RoutesBuilderPage").
					WithLayer("p_website.public_asset", views.MethodLayer{
						Method:  http.MethodGet,
						Handler: publicAssetHandler,
					}),
			},
		},
	}
}
