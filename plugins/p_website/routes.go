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
		},
	}
}
