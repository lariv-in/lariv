package p_contacts

import (
	"github.com/lariv-in/lago/lago"
)

func registerRoutes() {
	_ = lago.RegistryRoute.Register("contacts.DefaultRoute", lago.Route{
		Path:    AppUrl,
		Handler: lago.NewDynamicView("contacts.ListView"),
	})

	_ = lago.RegistryRoute.Register("contacts.CreateRoute", lago.Route{
		Path:    AppUrl + "create/",
		Handler: lago.NewDynamicView("contacts.CreateView"),
	})

	_ = lago.RegistryRoute.Register("contacts.DetailRoute", lago.Route{
		Path:    AppUrl + "{id}/",
		Handler: lago.NewDynamicView("contacts.DetailView"),
	})

	_ = lago.RegistryRoute.Register("contacts.UpdateRoute", lago.Route{
		Path:    AppUrl + "{id}/edit/",
		Handler: lago.NewDynamicView("contacts.UpdateView"),
	})

	_ = lago.RegistryRoute.Register("contacts.DeleteRoute", lago.Route{
		Path:    AppUrl + "{id}/delete/",
		Handler: lago.NewDynamicView("contacts.DeleteView"),
	})

	_ = lago.RegistryRoute.Register("contacts.SelectRoute", lago.Route{
		Path:    AppUrl + "select/",
		Handler: lago.NewDynamicView("contacts.SelectView"),
	})
}

func init() {
	registerRoutes()
}
