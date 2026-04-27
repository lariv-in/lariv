package p_semesters

import "github.com/lariv-in/lago/lago"

func registerRoutes() {
	_ = lago.RegistryRoute.Register("semesters.DefaultRoute", lago.Route{Path: AppURL, Handler: lago.NewDynamicView("semesters.ListView")})
	_ = lago.RegistryRoute.Register("semesters.CreateRoute", lago.Route{Path: AppURL + "create/", Handler: lago.NewDynamicView("semesters.CreateView")})
	_ = lago.RegistryRoute.Register("semesters.DetailRoute", lago.Route{Path: AppURL + "{id}/", Handler: lago.NewDynamicView("semesters.DetailView")})
	_ = lago.RegistryRoute.Register("semesters.UpdateRoute", lago.Route{Path: AppURL + "{id}/edit/", Handler: lago.NewDynamicView("semesters.UpdateView")})
	_ = lago.RegistryRoute.Register("semesters.DeleteRoute", lago.Route{Path: AppURL + "{id}/delete/", Handler: lago.NewDynamicView("semesters.DeleteView")})
	_ = lago.RegistryRoute.Register("semesters.SelectRoute", lago.Route{Path: AppURL + "select/", Handler: lago.NewDynamicView("semesters.SelectView")})
}

func init() { registerRoutes() }
