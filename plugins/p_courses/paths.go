package p_courses

import (
	"github.com/lariv-in/lago"
)

func registerRoutes() {
	_ = lago.RegistryRoute.Register("courses.DefaultRoute", lago.Route{
		Path:    AppUrl,
		Handler: lago.NewDynamicView("courses.ListView"),
	})

	_ = lago.RegistryRoute.Register("courses.CreateRoute", lago.Route{
		Path:    AppUrl + "create/",
		Handler: lago.NewDynamicView("courses.CreateView"),
	})

	_ = lago.RegistryRoute.Register("courses.DetailRoute", lago.Route{
		Path:    AppUrl + "{id}/",
		Handler: lago.NewDynamicView("courses.DetailView"),
	})

	_ = lago.RegistryRoute.Register("courses.UpdateRoute", lago.Route{
		Path:    AppUrl + "{id}/edit/",
		Handler: lago.NewDynamicView("courses.UpdateView"),
	})

	_ = lago.RegistryRoute.Register("courses.DeleteRoute", lago.Route{
		Path:    AppUrl + "{id}/delete/",
		Handler: lago.NewDynamicView("courses.DeleteView"),
	})

	_ = lago.RegistryRoute.Register("courses.SelectRoute", lago.Route{
		Path:    AppUrl + "select/",
		Handler: lago.NewDynamicView("courses.SelectView"),
	})

	_ = lago.RegistryRoute.Register("courses.MultiSelectRoute", lago.Route{
		Path:    AppUrl + "multi-select/",
		Handler: lago.NewDynamicView("courses.MultiSelectView"),
	})
}

func init() {
	registerRoutes()
}

