package p_announcements

import "github.com/lariv-in/lago/lago"

func registerRoutes() {
	_ = lago.RegistryRoute.Register("announcements.DefaultRoute", lago.Route{Path: AppURL, Handler: lago.NewDynamicView("announcements.ListView")})
	_ = lago.RegistryRoute.Register("announcements.CreateRoute", lago.Route{Path: AppURL + "create/", Handler: lago.NewDynamicView("announcements.CreateView")})
	_ = lago.RegistryRoute.Register("announcements.DetailRoute", lago.Route{Path: AppURL + "{id}/", Handler: lago.NewDynamicView("announcements.DetailView")})
	_ = lago.RegistryRoute.Register("announcements.UpdateRoute", lago.Route{Path: AppURL + "{id}/edit/", Handler: lago.NewDynamicView("announcements.UpdateView")})
	_ = lago.RegistryRoute.Register("announcements.DeleteRoute", lago.Route{Path: AppURL + "{id}/delete/", Handler: lago.NewDynamicView("announcements.DeleteView")})
}

func init() { registerRoutes() }
