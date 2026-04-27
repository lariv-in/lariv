package p_assignments

import "github.com/lariv-in/lago/lago"

func registerRoutes() {
	_ = lago.RegistryRoute.Register("assignments.DefaultRoute", lago.Route{Path: AppURL, Handler: lago.NewDynamicView("assignments.ListView")})
	_ = lago.RegistryRoute.Register("assignments.CreateRoute", lago.Route{Path: AppURL + "create/", Handler: lago.NewDynamicView("assignments.CreateView")})
	_ = lago.RegistryRoute.Register("assignments.DetailRoute", lago.Route{Path: AppURL + "{id}/", Handler: lago.NewDynamicView("assignments.DetailView")})
	_ = lago.RegistryRoute.Register("assignments.UpdateRoute", lago.Route{Path: AppURL + "{id}/edit/", Handler: lago.NewDynamicView("assignments.UpdateView")})
	_ = lago.RegistryRoute.Register("assignments.DeleteRoute", lago.Route{Path: AppURL + "{id}/delete/", Handler: lago.NewDynamicView("assignments.DeleteView")})
}

func init() { registerRoutes() }
