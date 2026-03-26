package p_nirmancampus_assignments

import (
	"github.com/lariv-in/lago/lago"
)

func registerRoutes() {
	_ = lago.RegistryRoute.Register("assignments.DefaultRoute", lago.Route{
		Path:    AppUrl,
		Handler: lago.NewDynamicView("assignments.ListView"),
	})

	_ = lago.RegistryRoute.Register("assignments.CreateRoute", lago.Route{
		Path:    AppUrl + "create/",
		Handler: lago.NewDynamicView("assignments.CreateView"),
	})

	_ = lago.RegistryRoute.Register("assignments.DetailRoute", lago.Route{
		Path:    AppUrl + "{id}/",
		Handler: lago.NewDynamicView("assignments.DetailView"),
	})

	_ = lago.RegistryRoute.Register("assignments.UpdateRoute", lago.Route{
		Path:    AppUrl + "{id}/edit/",
		Handler: lago.NewDynamicView("assignments.UpdateView"),
	})

	_ = lago.RegistryRoute.Register("assignments.DeleteRoute", lago.Route{
		Path:    AppUrl + "{id}/delete/",
		Handler: lago.NewDynamicView("assignments.DeleteView"),
	})

	_ = lago.RegistryRoute.Register("assignments.SelectRoute", lago.Route{
		Path:    AppUrl + "select/",
		Handler: lago.NewDynamicView("assignments.SelectView"),
	})
}

func init() {
	registerRoutes()
}
