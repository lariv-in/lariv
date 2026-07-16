package p_export

import (
	"github.com/lariv-in/lariv"
	"github.com/lariv-in/lariv/registry"
)

func pluginRoutes() lariv.PluginFeatures[lariv.Route] {
	return lariv.PluginFeatures[lariv.Route]{
		Entries: []registry.Pair[string, lariv.Route]{
			{Key: "export.PageRoute", Value: lariv.Route{
				Path:    AppUrl,
				Handler: lariv.NewDynamicView("export.PageView"),
			}},
			{Key: "export.DownloadRoute", Value: lariv.Route{
				Path:    AppUrl + "download/",
				Handler: lariv.NewDynamicView("export.DownloadView"),
			}},
		},
	}
}
