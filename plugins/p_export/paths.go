package p_export

import (
	"github.com/lariv-in/lago"
	"github.com/lariv-in/lago/registry"
)

func pluginRoutes() lago.PluginFeatures[lago.Route] {
	return lago.PluginFeatures[lago.Route]{
		Entries: []registry.Pair[string, lago.Route]{
			{Key: "export.PageRoute", Value: lago.Route{
				Path:    AppUrl,
				Handler: lago.NewDynamicView("export.PageView"),
			}},
			{Key: "export.DownloadRoute", Value: lago.Route{
				Path:    AppUrl + "download/",
				Handler: lago.NewDynamicView("export.DownloadView"),
			}},
		},
	}
}
