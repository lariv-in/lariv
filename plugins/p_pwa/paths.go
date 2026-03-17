package p_pwa

import "github.com/lariv-in/lago"

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
}

