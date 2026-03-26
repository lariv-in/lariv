package p_nirmancampus_announcements

import (
	"github.com/lariv-in/lago/lago"
)

func registerRoutes() {
	_ = lago.RegistryRoute.Register("announcements.DefaultRoute", lago.Route{
		Path:    AppUrl,
		Handler: lago.NewDynamicView("announcements.ListView"),
	})

	_ = lago.RegistryRoute.Register("announcements.CreateRoute", lago.Route{
		Path:    AppUrl + "create/",
		Handler: lago.NewDynamicView("announcements.CreateView"),
	})

	_ = lago.RegistryRoute.Register("announcements.DetailRoute", lago.Route{
		Path:    AppUrl + "{id}/",
		Handler: lago.NewDynamicView("announcements.DetailView"),
	})

	_ = lago.RegistryRoute.Register("announcements.UpdateRoute", lago.Route{
		Path:    AppUrl + "{id}/edit/",
		Handler: lago.NewDynamicView("announcements.UpdateView"),
	})

	_ = lago.RegistryRoute.Register("announcements.DeleteRoute", lago.Route{
		Path:    AppUrl + "{id}/delete/",
		Handler: lago.NewDynamicView("announcements.DeleteView"),
	})

	_ = lago.RegistryRoute.Register("announcements.SelectRoute", lago.Route{
		Path:    AppUrl + "select/",
		Handler: lago.NewDynamicView("announcements.SelectView"),
	})
}

func init() {
	registerRoutes()
}
