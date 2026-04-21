package p_seer_opensky

import "github.com/lariv-in/lago/lago"

func registerRoutes() {
	_ = lago.RegistryRoute.Register("seer_opensky.DefaultRoute", lago.Route{
		Path:    AppUrl,
		Handler: lago.NewDynamicView("seer_opensky.MapView"),
	})

	_ = lago.RegistryRoute.Register("seer_opensky.MapRoute", lago.Route{
		Path:    AppUrl + "map/",
		Handler: lago.NewDynamicView("seer_opensky.MapView"),
	})

	registerStatesAPIRoute()
}

func init() {
	registerRoutes()
}
