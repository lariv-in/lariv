package p_programs

import "github.com/lariv-in/lago/lago"

func registerRoutes() {
	_ = lago.RegistryRoute.Register("programs.DefaultRoute", lago.Route{Path: AppURL, Handler: lago.NewDynamicView("programs.ListView")})
	_ = lago.RegistryRoute.Register("programs.CreateRoute", lago.Route{Path: AppURL + "create/", Handler: lago.NewDynamicView("programs.CreateView")})
	_ = lago.RegistryRoute.Register("programs.DetailRoute", lago.Route{Path: AppURL + "{id}/", Handler: lago.NewDynamicView("programs.DetailView")})
	_ = lago.RegistryRoute.Register("programs.UpdateRoute", lago.Route{Path: AppURL + "{id}/edit/", Handler: lago.NewDynamicView("programs.UpdateView")})
	_ = lago.RegistryRoute.Register("programs.DeleteRoute", lago.Route{Path: AppURL + "{id}/delete/", Handler: lago.NewDynamicView("programs.DeleteView")})
	_ = lago.RegistryRoute.Register("programs.SelectRoute", lago.Route{Path: AppURL + "select/", Handler: lago.NewDynamicView("programs.SelectView")})
	_ = lago.RegistryRoute.Register("programs.MultiSelectRoute", lago.Route{Path: AppURL + "multi-select/", Handler: lago.NewDynamicView("programs.MultiSelectView")})
}

func init() { registerRoutes() }
