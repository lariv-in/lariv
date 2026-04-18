package p_seer_intel

import (
	"github.com/lariv-in/lago/lago"
)

func registerRoutes() {
	_ = lago.RegistryRoute.Register("seer_intel.DefaultRoute", lago.Route{
		Path:    AppUrl,
		Handler: lago.NewDynamicView("seer_intel.ListView"),
	})

	_ = lago.RegistryRoute.Register("seer_intel.DetailRoute", lago.Route{
		Path:    AppUrl + "{id}/",
		Handler: lago.NewDynamicView("seer_intel.DetailView"),
	})
}

func init() {
	registerRoutes()
}
