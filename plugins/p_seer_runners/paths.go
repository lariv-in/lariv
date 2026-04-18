package p_seer_runners

import (
	"github.com/lariv-in/lago/lago"
)

func registerRoutes() {
	_ = lago.RegistryRoute.Register("seer_runners.DefaultRoute", lago.Route{
		Path:    AppUrl,
		Handler: lago.NewDynamicView("seer_runners.ListView"),
	})

	_ = lago.RegistryRoute.Register("seer_runners.CreateRoute", lago.Route{
		Path:    AppUrl + "create/",
		Handler: lago.NewDynamicView("seer_runners.CreateView"),
	})

	_ = lago.RegistryRoute.Register("seer_runners.DetailRoute", lago.Route{
		Path:    AppUrl + "{id}/",
		Handler: lago.NewDynamicView("seer_runners.DetailView"),
	})
}

func init() {
	registerRoutes()
}
