package p_students

import "github.com/lariv-in/lago/lago"

func registerRoutes() {
	_ = lago.RegistryRoute.Register("students.DefaultRoute", lago.Route{
		Path:    AppURL,
		Handler: lago.NewDynamicView("students.ListView"),
	})

	_ = lago.RegistryRoute.Register("students.CreateRoute", lago.Route{
		Path:    AppURL + "create/",
		Handler: lago.NewDynamicView("students.CreateView"),
	})

	_ = lago.RegistryRoute.Register("students.DetailRoute", lago.Route{
		Path:    AppURL + "{id}/",
		Handler: lago.NewDynamicView("students.DetailView"),
	})

	_ = lago.RegistryRoute.Register("students.UpdateRoute", lago.Route{
		Path:    AppURL + "{id}/edit/",
		Handler: lago.NewDynamicView("students.UpdateView"),
	})

	_ = lago.RegistryRoute.Register("students.DeleteRoute", lago.Route{
		Path:    AppURL + "{id}/delete/",
		Handler: lago.NewDynamicView("students.DeleteView"),
	})
	_ = lago.RegistryRoute.Register("students.SelectRoute", lago.Route{
		Path:    AppURL + "select/",
		Handler: lago.NewDynamicView("students.SelectView"),
	})
	_ = lago.RegistryRoute.Register("students.MultiSelectRoute", lago.Route{
		Path:    AppURL + "multi-select/",
		Handler: lago.NewDynamicView("students.MultiSelectView"),
	})
}

func init() {
	registerRoutes()
}
