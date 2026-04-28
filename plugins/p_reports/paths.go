package p_reports

import "github.com/lariv-in/lago/lago"

func registerRoutes() {
	_ = lago.RegistryRoute.Register("reports.DefaultRoute", lago.Route{Path: AppURL, Handler: lago.NewDynamicView("reports.ListView")})
	_ = lago.RegistryRoute.Register("reports.CreateRoute", lago.Route{Path: AppURL + "create/", Handler: lago.NewDynamicView("reports.CreateView")})
	_ = lago.RegistryRoute.Register("reports.DetailRoute", lago.Route{Path: AppURL + "{id}/", Handler: lago.NewDynamicView("reports.DetailView")})
	_ = lago.RegistryRoute.Register("reports.UpdateRoute", lago.Route{Path: AppURL + "{id}/edit/", Handler: lago.NewDynamicView("reports.UpdateView")})
	_ = lago.RegistryRoute.Register("reports.DeleteRoute", lago.Route{Path: AppURL + "{id}/delete/", Handler: lago.NewDynamicView("reports.DeleteView")})
}

func init() { registerRoutes() }
