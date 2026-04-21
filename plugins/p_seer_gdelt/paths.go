package p_seer_gdelt

import "github.com/lariv-in/lago/lago"

func registerRoutes() {
	_ = lago.RegistryRoute.Register("seer_gdelt.DefaultRoute", lago.Route{
		Path:    AppUrl,
		Handler: lago.NewDynamicView("seer_gdelt.SearchView"),
	})

	_ = lago.RegistryRoute.Register("seer_gdelt.SearchRoute", lago.Route{
		Path:    AppUrl + "search/",
		Handler: lago.NewDynamicView("seer_gdelt.SearchView"),
	})

	_ = lago.RegistryRoute.Register("seer_gdelt.EventListRoute", lago.Route{
		Path:    AppUrl + "events/",
		Handler: lago.NewDynamicView("seer_gdelt.EventListView"),
	})

	_ = lago.RegistryRoute.Register("seer_gdelt.EventCreateRoute", lago.Route{
		Path:    AppUrl + "events/create/",
		Handler: lago.NewDynamicView("seer_gdelt.EventCreateView"),
	})

	_ = lago.RegistryRoute.Register("seer_gdelt.EventDetailRoute", lago.Route{
		Path:    AppUrl + "events/{id}/",
		Handler: lago.NewDynamicView("seer_gdelt.EventDetailView"),
	})

	_ = lago.RegistryRoute.Register("seer_gdelt.EventUpdateRoute", lago.Route{
		Path:    AppUrl + "events/{id}/edit/",
		Handler: lago.NewDynamicView("seer_gdelt.EventUpdateView"),
	})

	_ = lago.RegistryRoute.Register("seer_gdelt.EventDeleteRoute", lago.Route{
		Path:    AppUrl + "events/{id}/delete/",
		Handler: lago.NewDynamicView("seer_gdelt.EventDeleteView"),
	})

	_ = lago.RegistryRoute.Register("seer_gdelt.MapRoute", lago.Route{
		Path:    AppUrl + "map/",
		Handler: lago.NewDynamicView("seer_gdelt.MapView"),
	})
}

func init() {
	registerRoutes()
}
