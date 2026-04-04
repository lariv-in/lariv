package p_nirmancampus_students

import (
	"github.com/lariv-in/lago/lago"
)

func registerRoutes() {
	_ = lago.RegistryRoute.Register("students.DefaultRoute", lago.Route{
		Path:    AppUrl,
		Handler: lago.NewDynamicView("students.ListView"),
	})

	_ = lago.RegistryRoute.Register("students.CreateRoute", lago.Route{
		Path:    AppUrl + "create/",
		Handler: lago.NewDynamicView("students.CreateView"),
	})

	_ = lago.RegistryRoute.Register("students.DetailRoute", lago.Route{
		Path:    AppUrl + "{id}/",
		Handler: lago.NewDynamicView("students.DetailView"),
	})

	_ = lago.RegistryRoute.Register("students.UpdateRoute", lago.Route{
		Path:    AppUrl + "{id}/edit/",
		Handler: lago.NewDynamicView("students.UpdateView"),
	})

	_ = lago.RegistryRoute.Register("students.DeleteRoute", lago.Route{
		Path:    AppUrl + "{id}/delete/",
		Handler: lago.NewDynamicView("students.DeleteView"),
	})

	_ = lago.RegistryRoute.Register("students.SelectRoute", lago.Route{
		Path:    AppUrl + "select/",
		Handler: lago.NewDynamicView("students.SelectView"),
	})

	_ = lago.RegistryRoute.Register("students.UserPickRoute", lago.Route{
		Path:    AppUrl + "addon/pick-user/",
		Handler: lago.NewDynamicView("students.UserPickView"),
	})
}

func init() {
	registerRoutes()
}
