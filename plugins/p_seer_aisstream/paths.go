package p_seer_aisstream

import (
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
)

func registerRoutes() {
	_ = lago.RegistryRoute.Register("seer_aisstream.DefaultRoute", lago.Route{
		Path:    AppUrl,
		Handler: lago.NewDynamicView("seer_aisstream.MessageListView"),
	})
	_ = lago.RegistryRoute.Register("seer_aisstream.MessageListRoute", lago.Route{
		Path:    AppUrl + "messages/",
		Handler: lago.NewDynamicView("seer_aisstream.MessageListView"),
	})
	_ = lago.RegistryRoute.Register("seer_aisstream.MessageDetailRoute", lago.Route{
		Path:    AppUrl + "messages/{id}/",
		Handler: lago.NewDynamicView("seer_aisstream.MessageDetailView"),
	})
	_ = lago.RegistryRoute.Register("seer_aisstream.MapRouteUnderMessages", lago.Route{
		Path:    AppUrl + "messages/map/",
		Handler: lago.NewDynamicView("seer_aisstream.MapView"),
	})
	_ = lago.RegistryRoute.Register("seer_aisstream.MapRoute", lago.Route{
		Path:    AppUrl + "map/",
		Handler: lago.NewDynamicView("seer_aisstream.MapView"),
	})
	_ = lago.RegistryRoute.Register("seer_aisstream.MapDataRoute", lago.Route{
		Path:    AppUrl + "map/data/",
		Handler: p_users.RequireAuth(aisStreamMapDataHandler{}),
	})
}

func init() {
	registerRoutes()
}
