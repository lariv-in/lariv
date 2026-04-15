package p_nirmancampus_studentpayments

import "github.com/lariv-in/lago/lago"

func registerRoutes() {
	_ = lago.RegistryRoute.Register("studentpayments.DefaultRoute", lago.Route{
		Path:    AppUrl,
		Handler: lago.NewDynamicView("studentpayments.ListView"),
	})

	_ = lago.RegistryRoute.Register("studentpayments.CreateRoute", lago.Route{
		Path:    AppUrl + "create/",
		Handler: lago.NewDynamicView("studentpayments.CreateView"),
	})

	_ = lago.RegistryRoute.Register("studentpayments.DetailRoute", lago.Route{
		Path:    AppUrl + "{id}/",
		Handler: lago.NewDynamicView("studentpayments.DetailView"),
	})

	_ = lago.RegistryRoute.Register("studentpayments.UpdateRoute", lago.Route{
		Path:    AppUrl + "{id}/edit/",
		Handler: lago.NewDynamicView("studentpayments.UpdateView"),
	})

	_ = lago.RegistryRoute.Register("studentpayments.DeleteRoute", lago.Route{
		Path:    AppUrl + "{id}/delete/",
		Handler: lago.NewDynamicView("studentpayments.DeleteView"),
	})
}

func init() {
	registerRoutes()
}
