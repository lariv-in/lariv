package p_dashboard

import (
	"github.com/lariv-in/lariv"
	"github.com/lariv-in/lariv/registry"
)

// pluginRoutes serves the apps grid at [AppUrl]. Registry key matches
// [lariv.RoutePath]("dashboard.AppsPage", …) used across menu and redirects.
func pluginRoutes() lariv.PluginFeatures[lariv.Route] {
	return lariv.PluginFeatures[lariv.Route]{
		Entries: []registry.Pair[string, lariv.Route]{
			{Key: "dashboard.AppsPage", Value: lariv.Route{
				Path:    AppUrl,
				Handler: lariv.NewDynamicView("dashboard.AppsView"),
			}},
		},
	}
}
