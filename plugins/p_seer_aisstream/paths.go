package p_seer_aisstream

import "github.com/lariv-in/lago/lago"

func registerRoutes() {
	_ = lago.RegistryRoute.Register("seer_aisstream.DefaultRoute", lago.Route{
		Path:    AppUrl,
		Handler: lago.NewDynamicView("seer_aisstream.MapView"),
	})

	_ = lago.RegistryRoute.Register("seer_aisstream.MapRoute", lago.Route{
		Path:    AppUrl + "map/",
		Handler: lago.NewDynamicView("seer_aisstream.MapView"),
	})

	registerVesselsAPIRoute()
}

func init() {
	registerRoutes()
}
