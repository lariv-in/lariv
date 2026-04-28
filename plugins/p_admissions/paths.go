package p_admissions

import "github.com/lariv-in/lago/lago"

func registerRoutes() {
	_ = lago.RegistryRoute.Register("admissions.DefaultRoute", lago.Route{Path: AppURL, Handler: lago.NewDynamicView("admissions.ListView")})
	_ = lago.RegistryRoute.Register("admissions.CreateRoute", lago.Route{Path: AppURL + "create/", Handler: lago.NewDynamicView("admissions.CreateView")})
	_ = lago.RegistryRoute.Register("admissions.DetailRoute", lago.Route{Path: AppURL + "{id}/", Handler: lago.NewDynamicView("admissions.DetailView")})
	_ = lago.RegistryRoute.Register("admissions.UpdateRoute", lago.Route{Path: AppURL + "{id}/edit/", Handler: lago.NewDynamicView("admissions.UpdateView")})
	_ = lago.RegistryRoute.Register("admissions.DeleteRoute", lago.Route{Path: AppURL + "{id}/delete/", Handler: lago.NewDynamicView("admissions.DeleteView")})
}

func init() { registerRoutes() }
