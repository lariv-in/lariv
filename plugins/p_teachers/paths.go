package p_teachers

import "github.com/lariv-in/lago/lago"

func registerRoutes() {
	_ = lago.RegistryRoute.Register("teachers.DefaultRoute", lago.Route{
		Path:    AppURL,
		Handler: lago.NewDynamicView("teachers.ListView"),
	})
	_ = lago.RegistryRoute.Register("teachers.CreateRoute", lago.Route{
		Path:    AppURL + "create/",
		Handler: lago.NewDynamicView("teachers.CreateView"),
	})
	_ = lago.RegistryRoute.Register("teachers.DetailRoute", lago.Route{
		Path:    AppURL + "{id}/",
		Handler: lago.NewDynamicView("teachers.DetailView"),
	})
	_ = lago.RegistryRoute.Register("teachers.UpdateRoute", lago.Route{
		Path:    AppURL + "{id}/edit/",
		Handler: lago.NewDynamicView("teachers.UpdateView"),
	})
	_ = lago.RegistryRoute.Register("teachers.DeleteRoute", lago.Route{
		Path:    AppURL + "{id}/delete/",
		Handler: lago.NewDynamicView("teachers.DeleteView"),
	})
	_ = lago.RegistryRoute.Register("teachers.SelectRoute", lago.Route{
		Path:    AppURL + "select/",
		Handler: lago.NewDynamicView("teachers.SelectView"),
	})
	_ = lago.RegistryRoute.Register("teachers.MultiSelectRoute", lago.Route{
		Path:    AppURL + "multi-select/",
		Handler: lago.NewDynamicView("teachers.MultiSelectView"),
	})
}

func init() { registerRoutes() }
