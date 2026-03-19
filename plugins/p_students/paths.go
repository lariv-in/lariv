package p_students

import (
	"github.com/lariv-in/lago"
)

func registerRoutes() {
	// Default route - student list
	_ = lago.RegistryRoute.Register("students.DefaultRoute", lago.Route{
		Path:    AppUrl,
		Handler: lago.NewDynamicView("students.ListView"),
	})

	// Create route
	_ = lago.RegistryRoute.Register("students.CreateRoute", lago.Route{
		Path:    AppUrl + "create/",
		Handler: lago.NewDynamicView("students.CreateView"),
	})

	// Detail route
	_ = lago.RegistryRoute.Register("students.DetailRoute", lago.Route{
		Path:    AppUrl + "{id}/",
		Handler: lago.NewDynamicView("students.DetailView"),
	})

	// Update route (edit)
	_ = lago.RegistryRoute.Register("students.UpdateRoute", lago.Route{
		Path:    AppUrl + "{id}/edit/",
		Handler: lago.NewDynamicView("students.UpdateView"),
	})

	// Delete route
	_ = lago.RegistryRoute.Register("students.DeleteRoute", lago.Route{
		Path:    AppUrl + "{id}/delete/",
		Handler: lago.NewDynamicView("students.DeleteView"),
	})

	// Select route - for foreign key selection modal
	_ = lago.RegistryRoute.Register("students.SelectRoute", lago.Route{
		Path:    AppUrl + "select/",
		Handler: lago.NewDynamicView("students.SelectView"),
	})
}

func init() {
	registerRoutes()
}
