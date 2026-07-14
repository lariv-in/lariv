package p_pwa

import (
	"github.com/lariv-in/lago"
	"github.com/lariv-in/lago/registry"
)

func pluginRoutes() lago.PluginFeatures[lago.Route] {
	return lago.PluginFeatures[lago.Route]{
		Entries: []registry.Pair[string, lago.Route]{
			{
				Key: "pwa.ManifestRoute",
				Value: lago.Route{
					Path:    "/app.webmanifest",
					Handler: lago.NewDynamicView(manifestViewKey),
				},
			},
			{
				Key: "pwa.ServiceWorkerRoute",
				Value: lago.Route{
					Path:    "/serviceworker.js",
					Handler: lago.NewDynamicView(serviceWorkerViewKey),
				},
			},
			{
				Key: "pwa.OfflineRoute",
				Value: lago.Route{
					Path:    "/offline",
					Handler: lago.NewDynamicView(offlineViewKey),
				},
			},
			{
				Key: "pwa.assetLinks",
				Value: lago.Route{
					Path:    "/.well-known/assetlinks.json",
					Handler: lago.NewDynamicView(assetLinksViewKey),
				},
			},
			{
				Key: "pwa.StaticPwaBaseRoute",
				Value: lago.Route{
					Path:    "/static/pwa/",
					Handler: lago.NewDynamicView(staticPwaViewKey),
				},
			},
			{
				Key: "pwa.StaticPwaFilesRoute",
				Value: lago.Route{
					Path:    "/static/pwa/{path...}",
					Handler: lago.NewDynamicView(staticPwaViewKey),
				},
			},
		},
	}
}
