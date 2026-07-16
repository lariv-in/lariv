package p_pwa

import (
	"github.com/lariv-in/lariv"
	"github.com/lariv-in/lariv/registry"
)

func pluginRoutes() lariv.PluginFeatures[lariv.Route] {
	return lariv.PluginFeatures[lariv.Route]{
		Entries: []registry.Pair[string, lariv.Route]{
			{
				Key: "pwa.ManifestRoute",
				Value: lariv.Route{
					Path:    "/app.webmanifest",
					Handler: lariv.NewDynamicView(manifestViewKey),
				},
			},
			{
				Key: "pwa.ServiceWorkerRoute",
				Value: lariv.Route{
					Path:    "/serviceworker.js",
					Handler: lariv.NewDynamicView(serviceWorkerViewKey),
				},
			},
			{
				Key: "pwa.OfflineRoute",
				Value: lariv.Route{
					Path:    "/offline",
					Handler: lariv.NewDynamicView(offlineViewKey),
				},
			},
			{
				Key: "pwa.assetLinks",
				Value: lariv.Route{
					Path:    "/.well-known/assetlinks.json",
					Handler: lariv.NewDynamicView(assetLinksViewKey),
				},
			},
			{
				Key: "pwa.StaticPwaBaseRoute",
				Value: lariv.Route{
					Path:    "/static/pwa/",
					Handler: lariv.NewDynamicView(staticPwaViewKey),
				},
			},
			{
				Key: "pwa.StaticPwaFilesRoute",
				Value: lariv.Route{
					Path:    "/static/pwa/{path...}",
					Handler: lariv.NewDynamicView(staticPwaViewKey),
				},
			},
		},
	}
}
