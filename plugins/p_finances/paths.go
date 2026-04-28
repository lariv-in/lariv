package p_finances

import "github.com/lariv-in/lago/lago"

func registerRoutes() {
	_ = lago.RegistryRoute.Register("finances.DefaultRoute", lago.Route{Path: AppURL, Handler: lago.NewDynamicView("finances.ListView")})
	_ = lago.RegistryRoute.Register("finances.CreateRoute", lago.Route{Path: AppURL + "create/", Handler: lago.NewDynamicView("finances.CreateView")})
	_ = lago.RegistryRoute.Register("finances.DetailRoute", lago.Route{Path: AppURL + "{id}/", Handler: lago.NewDynamicView("finances.DetailView")})
	_ = lago.RegistryRoute.Register("finances.UpdateRoute", lago.Route{Path: AppURL + "{id}/edit/", Handler: lago.NewDynamicView("finances.UpdateView")})
	_ = lago.RegistryRoute.Register("finances.DeleteRoute", lago.Route{Path: AppURL + "{id}/delete/", Handler: lago.NewDynamicView("finances.DeleteView")})
}

func init() { registerRoutes() }
