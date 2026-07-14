package p_dashboard

import (
	"github.com/lariv-in/lago"
	"github.com/lariv-in/lago/registry"
)

// pluginRoutes serves the apps grid at [AppUrl]. Registry key matches
// [lago.RoutePath]("dashboard.AppsPage", …) used across menu and redirects.
func pluginRoutes() lago.PluginFeatures[lago.Route] {
	return lago.PluginFeatures[lago.Route]{
		Entries: []registry.Pair[string, lago.Route]{
			{Key: "dashboard.AppsPage", Value: lago.Route{
				Path:    AppUrl,
				Handler: lago.NewDynamicView("dashboard.AppsView"),
			}},
		},
	}
}
