package p_timetable

import "github.com/lariv-in/lago/lago"

func registerRoutes() {
	_ = lago.RegistryRoute.Register("timetable.DefaultRoute", lago.Route{Path: AppURL, Handler: lago.NewDynamicView("timetable.ListView")})
	_ = lago.RegistryRoute.Register("timetable.CreateRoute", lago.Route{Path: AppURL + "create/", Handler: lago.NewDynamicView("timetable.CreateView")})
	_ = lago.RegistryRoute.Register("timetable.DetailRoute", lago.Route{Path: AppURL + "{id}/", Handler: lago.NewDynamicView("timetable.DetailView")})
	_ = lago.RegistryRoute.Register("timetable.UpdateRoute", lago.Route{Path: AppURL + "{id}/edit/", Handler: lago.NewDynamicView("timetable.UpdateView")})
	_ = lago.RegistryRoute.Register("timetable.DeleteRoute", lago.Route{Path: AppURL + "{id}/delete/", Handler: lago.NewDynamicView("timetable.DeleteView")})
}

func init() { registerRoutes() }
