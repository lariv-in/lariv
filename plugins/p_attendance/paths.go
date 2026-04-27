package p_attendance

import "github.com/lariv-in/lago/lago"

func registerRoutes() {
	_ = lago.RegistryRoute.Register("attendance.DefaultRoute", lago.Route{Path: AppURL, Handler: lago.NewDynamicView("attendance.ListView")})
	_ = lago.RegistryRoute.Register("attendance.CreateRoute", lago.Route{Path: AppURL + "create/", Handler: lago.NewDynamicView("attendance.CreateView")})
	_ = lago.RegistryRoute.Register("attendance.DetailRoute", lago.Route{Path: AppURL + "{id}/", Handler: lago.NewDynamicView("attendance.DetailView")})
	_ = lago.RegistryRoute.Register("attendance.UpdateRoute", lago.Route{Path: AppURL + "{id}/edit/", Handler: lago.NewDynamicView("attendance.UpdateView")})
	_ = lago.RegistryRoute.Register("attendance.DeleteRoute", lago.Route{Path: AppURL + "{id}/delete/", Handler: lago.NewDynamicView("attendance.DeleteView")})
}

func init() { registerRoutes() }
