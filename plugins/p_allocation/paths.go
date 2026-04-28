package p_allocation

import "github.com/lariv-in/lago/lago"

func registerRoutes() {
	_ = lago.RegistryRoute.Register("allocation.DefaultRoute", lago.Route{Path: AppURL, Handler: lago.NewDynamicView("allocation.ListView")})
	_ = lago.RegistryRoute.Register("allocation.CreateRoute", lago.Route{Path: AppURL + "create/", Handler: lago.NewDynamicView("allocation.CreateView")})
	_ = lago.RegistryRoute.Register("allocation.DetailRoute", lago.Route{Path: AppURL + "{id}/", Handler: lago.NewDynamicView("allocation.DetailView")})
	_ = lago.RegistryRoute.Register("allocation.UpdateRoute", lago.Route{Path: AppURL + "{id}/edit/", Handler: lago.NewDynamicView("allocation.UpdateView")})
	_ = lago.RegistryRoute.Register("allocation.DeleteRoute", lago.Route{Path: AppURL + "{id}/delete/", Handler: lago.NewDynamicView("allocation.DeleteView")})
}

func init() { registerRoutes() }
