package p_nirmancampus_studentapplications

import (
	"github.com/lariv-in/lago/lago"
)

func registerRoutes() {
	_ = lago.RegistryRoute.Register("studentapplications.DefaultRoute", lago.Route{
		Path:    AppUrl,
		Handler: lago.NewDynamicView("studentapplications.ListView"),
	})

	_ = lago.RegistryRoute.Register("studentapplications.CreateRoute", lago.Route{
		Path:    AppUrl + "create/",
		Handler: lago.NewDynamicView("studentapplications.CreateView"),
	})

	_ = lago.RegistryRoute.Register("studentapplications.PublicApplyRoute", lago.Route{
		Path:    AppUrl + "apply/",
		Handler: lago.NewDynamicView("studentapplications.PublicApplyView"),
	})

	_ = lago.RegistryRoute.Register("studentapplications.DetailRoute", lago.Route{
		Path:    AppUrl + "{id}/",
		Handler: lago.NewDynamicView("studentapplications.DetailView"),
	})

	_ = lago.RegistryRoute.Register("studentapplications.UpdateRoute", lago.Route{
		Path:    AppUrl + "{id}/edit/",
		Handler: lago.NewDynamicView("studentapplications.UpdateView"),
	})

	_ = lago.RegistryRoute.Register("studentapplications.DeleteRoute", lago.Route{
		Path:    AppUrl + "{id}/delete/",
		Handler: lago.NewDynamicView("studentapplications.DeleteView"),
	})
}

func init() {
	registerRoutes()
}
