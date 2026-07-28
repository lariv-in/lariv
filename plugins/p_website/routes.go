package p_website

import (
	"github.com/lariv-in/lariv"
	"github.com/lariv-in/lariv/registry"
)

func pluginRoutes() lariv.PluginFeatures[lariv.Route] {
	return lariv.PluginFeatures[lariv.Route]{
		Patches: []registry.Pair[string, func(lariv.Route) lariv.Route]{
			{
				Key: "core.HomeRoute",
				Value: func(old lariv.Route) lariv.Route {
					old.Path = "/{path...}"
					old.Handler = lariv.NewDynamicView("p_website.DynamicWebsiteView")
					return old
				},
			},
		},
		Entries: []registry.Pair[string, lariv.Route]{
			{
				Key: "p_website.RoutesListRoute",
				Value: lariv.Route{
					Path:    AppURL,
					Handler: lariv.NewDynamicView("p_website.RoutesListView"),
				},
			},
			{
				Key: "p_website.RoutesCreateRoute",
				Value: lariv.Route{
					Path:    AppURL + "create/",
					Handler: lariv.NewDynamicView("p_website.RoutesCreateView"),
				},
			},
			{
				Key: "p_website.RoutesDetailRoute",
				Value: lariv.Route{
					Path:    AppURL + "{id}/",
					Handler: lariv.NewDynamicView("p_website.RoutesDetailView"),
				},
			},
			{
				Key: "p_website.RoutesUpdateRoute",
				Value: lariv.Route{
					Path:    AppURL + "{id}/edit/",
					Handler: lariv.NewDynamicView("p_website.RoutesUpdateView"),
				},
			},
			{
				Key: "p_website.RoutesDeleteRoute",
				Value: lariv.Route{
					Path:    AppURL + "{id}/delete/",
					Handler: lariv.NewDynamicView("p_website.RoutesDeleteView"),
				},
			},
			{
				Key: "p_website.RoutesBuilderRoute",
				Value: lariv.Route{
					Path:    AppURL + "{id}/builder/",
					Handler: lariv.NewDynamicView("p_website.RoutesBuilderView"),
				},
			},
			{
				Key: "p_website.RoutesBuilderProjectRoute",
				Value: lariv.Route{
					Path:    AppURL + "{id}/builder/project/",
					Handler: lariv.NewDynamicView("p_website.RoutesBuilderProjectView"),
				},
			},
			{
				Key: "p_website.RoutesBuilderThemeRoute",
				Value: lariv.Route{
					Path:    AppURL + "{id}/builder/theme/",
					Handler: lariv.NewDynamicView("p_website.RoutesBuilderThemeView"),
				},
			},
			{
				Key: "p_website.BuilderAssetUploadRoute",
				Value: lariv.Route{
					Path:    AppURL + "builder/assets/",
					Handler: lariv.NewDynamicView("p_website.BuilderAssetUploadView"),
				},
			},
			{
				Key: "p_website.PublicAssetRoute",
				Value: lariv.Route{
					Path:    "/media/{id}/",
					Handler: lariv.NewDynamicView("p_website.PublicAssetView"),
				},
			},
		},
	}
}
