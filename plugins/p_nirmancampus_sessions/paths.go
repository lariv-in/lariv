package p_nirmancampus_sessions

import (
	"github.com/lariv-in/lago/lago"
)

func registerRoutes() {
	_ = lago.RegistryRoute.Register("sessions.DefaultRoute", lago.Route{
		Path:    AppUrl,
		Handler: lago.NewDynamicView("sessions.ListView"),
	})

	_ = lago.RegistryRoute.Register("sessions.CreateRoute", lago.Route{
		Path:    AppUrl + "create/",
		Handler: lago.NewDynamicView("sessions.CreateView"),
	})

	_ = lago.RegistryRoute.Register("sessions.DetailRoute", lago.Route{
		Path:    AppUrl + "{id}/",
		Handler: lago.NewDynamicView("sessions.DetailView"),
	})

	_ = lago.RegistryRoute.Register("sessions.UpdateRoute", lago.Route{
		Path:    AppUrl + "{id}/edit/",
		Handler: lago.NewDynamicView("sessions.UpdateView"),
	})

	_ = lago.RegistryRoute.Register("sessions.DeleteRoute", lago.Route{
		Path:    AppUrl + "{id}/delete/",
		Handler: lago.NewDynamicView("sessions.DeleteView"),
	})

	_ = lago.RegistryRoute.Register("sessions.SelectRoute", lago.Route{
		Path:    AppUrl + "select/",
		Handler: lago.NewDynamicView("sessions.SelectView"),
	})
}

func init() {
	registerRoutes()
}
