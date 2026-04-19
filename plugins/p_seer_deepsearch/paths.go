package p_seer_deepsearch

import "github.com/lariv-in/lago/lago"

func registerRoutes() {
	_ = lago.RegistryRoute.Register("seer_deepsearch.DefaultRoute", lago.Route{
		Path:    AppUrl,
		Handler: lago.NewDynamicView("seer_deepsearch.HomeView"),
	})

	_ = lago.RegistryRoute.Register("seer_deepsearch.StartRoute", lago.Route{
		Path:    AppUrl + "start/",
		Handler: lago.NewDynamicView("seer_deepsearch.StartView"),
	})

	// Literal "history/" before "{id}/" so /seer-deepsearch/history/ is not captured as an id segment.
	_ = lago.RegistryRoute.Register("seer_deepsearch.HistoryRoute", lago.Route{
		Path:    AppUrl + "history/",
		Handler: lago.NewDynamicView("seer_deepsearch.HistoryView"),
	})

	// Literal "{id}/stop/" and "{id}/restart/" before "{id}/" so they are not swallowed as id text.
	_ = lago.RegistryRoute.Register("seer_deepsearch.StopRoute", lago.Route{
		Path:    AppUrl + "{id}/stop/",
		Handler: lago.NewDynamicView("seer_deepsearch.StopView"),
	})
	_ = lago.RegistryRoute.Register("seer_deepsearch.RestartRoute", lago.Route{
		Path:    AppUrl + "{id}/restart/",
		Handler: lago.NewDynamicView("seer_deepsearch.RestartView"),
	})

	_ = lago.RegistryRoute.Register("seer_deepsearch.DetailRoute", lago.Route{
		Path:    AppUrl + "{id}/",
		Handler: lago.NewDynamicView("seer_deepsearch.DetailView"),
	})
}

func init() {
	registerRoutes()
}
