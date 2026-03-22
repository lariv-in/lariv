package p_programs

import (
	"github.com/lariv-in/lago/lago"
)

func registerRoutes() {
	_ = lago.RegistryRoute.Register("programs.DefaultRoute", lago.Route{
		Path:    AppUrl,
		Handler: lago.NewDynamicView("programs.ListView"),
	})

	_ = lago.RegistryRoute.Register("programs.CreateRoute", lago.Route{
		Path:    AppUrl + "create/",
		Handler: lago.NewDynamicView("programs.CreateView"),
	})

	_ = lago.RegistryRoute.Register("programs.DetailRoute", lago.Route{
		Path:    AppUrl + "{id}/",
		Handler: lago.NewDynamicView("programs.DetailView"),
	})

	_ = lago.RegistryRoute.Register("programs.UpdateRoute", lago.Route{
		Path:    AppUrl + "{id}/edit/",
		Handler: lago.NewDynamicView("programs.UpdateView"),
	})

	_ = lago.RegistryRoute.Register("programs.DeleteRoute", lago.Route{
		Path:    AppUrl + "{id}/delete/",
		Handler: lago.NewDynamicView("programs.DeleteView"),
	})

	_ = lago.RegistryRoute.Register("programs.SelectRoute", lago.Route{
		Path:    AppUrl + "select/",
		Handler: lago.NewDynamicView("programs.SelectView"),
	})
}

func init() {
	registerRoutes()
}

