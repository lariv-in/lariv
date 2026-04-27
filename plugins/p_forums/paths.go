package p_forums

import "github.com/lariv-in/lago/lago"

func registerRoutes() {
	_ = lago.RegistryRoute.Register("forums.DefaultRoute", lago.Route{Path: AppURL, Handler: lago.NewDynamicView("forums.ListView")})
	_ = lago.RegistryRoute.Register("forums.CreateRoute", lago.Route{Path: AppURL + "create/", Handler: lago.NewDynamicView("forums.CreateView")})
	_ = lago.RegistryRoute.Register("forums.DetailRoute", lago.Route{Path: AppURL + "{id}/", Handler: lago.NewDynamicView("forums.DetailView")})
	_ = lago.RegistryRoute.Register("forums.UpdateRoute", lago.Route{Path: AppURL + "{id}/edit/", Handler: lago.NewDynamicView("forums.UpdateView")})
	_ = lago.RegistryRoute.Register("forums.DeleteRoute", lago.Route{Path: AppURL + "{id}/delete/", Handler: lago.NewDynamicView("forums.DeleteView")})
}

func init() { registerRoutes() }
