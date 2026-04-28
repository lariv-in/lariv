package p_courses

import "github.com/lariv-in/lago/lago"

func registerRoutes() {
	_ = lago.RegistryRoute.Register("courses.DefaultRoute", lago.Route{Path: AppURL, Handler: lago.NewDynamicView("courses.ListView")})
	_ = lago.RegistryRoute.Register("courses.CreateRoute", lago.Route{Path: AppURL + "create/", Handler: lago.NewDynamicView("courses.CreateView")})
	_ = lago.RegistryRoute.Register("courses.DetailRoute", lago.Route{Path: AppURL + "{id}/", Handler: lago.NewDynamicView("courses.DetailView")})
	_ = lago.RegistryRoute.Register("courses.UpdateRoute", lago.Route{Path: AppURL + "{id}/edit/", Handler: lago.NewDynamicView("courses.UpdateView")})
	_ = lago.RegistryRoute.Register("courses.DeleteRoute", lago.Route{Path: AppURL + "{id}/delete/", Handler: lago.NewDynamicView("courses.DeleteView")})
	_ = lago.RegistryRoute.Register("courses.SelectRoute", lago.Route{Path: AppURL + "select/", Handler: lago.NewDynamicView("courses.SelectView")})
}

func init() { registerRoutes() }
