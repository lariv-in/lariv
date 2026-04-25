package p_seer_opensky

import (
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
)

func registerRoutes() {
	_ = lago.RegistryRoute.Register("seer_opensky.DefaultRoute", lago.Route{
		Path:    AppUrl,
		Handler: lago.NewDynamicView("seer_opensky.StateListView"),
	})
	_ = lago.RegistryRoute.Register("seer_opensky.StateListRoute", lago.Route{
		Path:    AppUrl + "states/",
		Handler: lago.NewDynamicView("seer_opensky.StateListView"),
	})
	_ = lago.RegistryRoute.Register("seer_opensky.MapRouteUnderStates", lago.Route{
		Path:    AppUrl + "states/map/",
		Handler: lago.NewDynamicView("seer_opensky.MapView"),
	})
	_ = lago.RegistryRoute.Register("seer_opensky.StateCreateRoute", lago.Route{
		Path:    AppUrl + "states/create/",
		Handler: lago.NewDynamicView("seer_opensky.StateCreateView"),
	})
	_ = lago.RegistryRoute.Register("seer_opensky.StateDetailRoute", lago.Route{
		Path:    AppUrl + "states/{id}/",
		Handler: lago.NewDynamicView("seer_opensky.StateDetailView"),
	})
	_ = lago.RegistryRoute.Register("seer_opensky.StateUpdateRoute", lago.Route{
		Path:    AppUrl + "states/{id}/edit/",
		Handler: lago.NewDynamicView("seer_opensky.StateUpdateView"),
	})
	_ = lago.RegistryRoute.Register("seer_opensky.StateDeleteRoute", lago.Route{
		Path:    AppUrl + "states/{id}/delete/",
		Handler: lago.NewDynamicView("seer_opensky.StateDeleteView"),
	})
	_ = lago.RegistryRoute.Register("seer_opensky.MapRoute", lago.Route{
		Path:    AppUrl + "map/",
		Handler: lago.NewDynamicView("seer_opensky.MapView"),
	})
	_ = lago.RegistryRoute.Register("seer_opensky.MapDataRoute", lago.Route{
		Path:    AppUrl + "map/data/",
		Handler: p_users.RequireAuth(openSkyMapDataHandler{}),
	})
}

func init() {
	registerRoutes()
}
