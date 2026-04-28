package p_events

import "github.com/lariv-in/lago/lago"

func registerRoutes() {
	_ = lago.RegistryRoute.Register("events.DefaultRoute", lago.Route{
		Path:    AppURL,
		Handler: lago.NewDynamicView("events.ListView"),
	})
	_ = lago.RegistryRoute.Register("events.CreateRoute", lago.Route{
		Path:    AppURL + "create/",
		Handler: lago.NewDynamicView("events.CreateView"),
	})
	_ = lago.RegistryRoute.Register("events.DetailRoute", lago.Route{
		Path:    AppURL + "{id}/",
		Handler: lago.NewDynamicView("events.DetailView"),
	})
	_ = lago.RegistryRoute.Register("events.UpdateRoute", lago.Route{
		Path:    AppURL + "{id}/edit/",
		Handler: lago.NewDynamicView("events.UpdateView"),
	})
	_ = lago.RegistryRoute.Register("events.DeleteRoute", lago.Route{
		Path:    AppURL + "{id}/delete/",
		Handler: lago.NewDynamicView("events.DeleteView"),
	})
}

func init() {
	registerRoutes()
}
