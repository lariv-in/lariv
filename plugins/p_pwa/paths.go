package p_pwa

import "github.com/lariv-in/lago/lago"

func init() {
	_ = lago.RegistryRoute.Register("pwa.ManifestRoute", lago.Route{
		Path:    "/app.webmanifest",
		Handler: lago.NewDynamicView(manifestViewKey),
	})

	_ = lago.RegistryRoute.Register("pwa.ServiceWorkerRoute", lago.Route{
		Path:    "/serviceworker.js",
		Handler: lago.NewDynamicView(serviceWorkerViewKey),
	})

	_ = lago.RegistryRoute.Register("pwa.OfflineRoute", lago.Route{
		Path:    "/offline",
		Handler: lago.NewDynamicView(offlineViewKey),
	})


	_ = lago.RegistryRoute.Register("pwa.assetLinks", lago.Route{
		Path:    "/.well-known/assetlinks.json",
		Handler: lago.NewDynamicView(assetLinksViewKey),
	})
	// Serve a filesystem directory under /static/pwa/.
	// Note: lago's router treats paths ending in "/" as exact matches, so we use
	// a wildcard pattern for nested file paths.
	_ = lago.RegistryRoute.Register("pwa.StaticPwaBaseRoute", lago.Route{
		Path:    "/static/pwa/",
		Handler: lago.NewDynamicView(staticPwaViewKey),
	})
	_ = lago.RegistryRoute.Register("pwa.StaticPwaFilesRoute", lago.Route{
		Path:    "/static/pwa/{path...}",
		Handler: lago.NewDynamicView(staticPwaViewKey),
	})
}
