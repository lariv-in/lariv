package p_teachers

import (
	"github.com/lariv-in/lago/lago"
)

func registerRoutes() {
	// Default route - teacher list
	_ = lago.RegistryRoute.Register("teachers.DefaultRoute", lago.Route{
		Path:    AppUrl,
		Handler: lago.NewDynamicView("teachers.ListView"),
	})

	// Create route
	_ = lago.RegistryRoute.Register("teachers.CreateRoute", lago.Route{
		Path:    AppUrl + "create/",
		Handler: lago.NewDynamicView("teachers.CreateView"),
	})

	// Detail route
	_ = lago.RegistryRoute.Register("teachers.DetailRoute", lago.Route{
		Path:    AppUrl + "{id}/",
		Handler: lago.NewDynamicView("teachers.DetailView"),
	})

	// Update route (edit)
	_ = lago.RegistryRoute.Register("teachers.UpdateRoute", lago.Route{
		Path:    AppUrl + "{id}/edit/",
		Handler: lago.NewDynamicView("teachers.UpdateView"),
	})

	// Delete route
	_ = lago.RegistryRoute.Register("teachers.DeleteRoute", lago.Route{
		Path:    AppUrl + "{id}/delete/",
		Handler: lago.NewDynamicView("teachers.DeleteView"),
	})

	// Select route - for foreign key selection modal
	_ = lago.RegistryRoute.Register("teachers.SelectRoute", lago.Route{
		Path:    AppUrl + "select/",
		Handler: lago.NewDynamicView("teachers.SelectView"),
	})

	// Multi-select route - for many-to-many selection modal
	_ = lago.RegistryRoute.Register("teachers.MultiSelectRoute", lago.Route{
		Path:    AppUrl + "multi-select/",
		Handler: lago.NewDynamicView("teachers.MultiSelectView"),
	})

}

func init() {
	registerRoutes()
}
