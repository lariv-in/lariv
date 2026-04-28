package p_syllabus

import "github.com/lariv-in/lago/lago"

func registerRoutes() {
	_ = lago.RegistryRoute.Register("syllabus.DefaultRoute", lago.Route{
		Path:    AppURL,
		Handler: lago.NewDynamicView("syllabus.ListView"),
	})
	_ = lago.RegistryRoute.Register("syllabus.CreateRoute", lago.Route{
		Path:    AppURL + "create/",
		Handler: lago.NewDynamicView("syllabus.CreateView"),
	})
	_ = lago.RegistryRoute.Register("syllabus.DetailRoute", lago.Route{
		Path:    AppURL + "{id}/",
		Handler: lago.NewDynamicView("syllabus.DetailView"),
	})
	_ = lago.RegistryRoute.Register("syllabus.UpdateRoute", lago.Route{
		Path:    AppURL + "{id}/edit/",
		Handler: lago.NewDynamicView("syllabus.UpdateView"),
	})
	_ = lago.RegistryRoute.Register("syllabus.DeleteRoute", lago.Route{
		Path:    AppURL + "{id}/delete/",
		Handler: lago.NewDynamicView("syllabus.DeleteView"),
	})
	_ = lago.RegistryRoute.Register("syllabus.MultiSelectRoute", lago.Route{
		Path:    AppURL + "multi-select/",
		Handler: lago.NewDynamicView("syllabus.MultiSelectView"),
	})
}

func init() {
	registerRoutes()
}
