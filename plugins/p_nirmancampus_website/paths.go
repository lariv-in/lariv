package p_nirmancampus_website

import (
	"github.com/lariv-in/lago/lago"
)

func init() {
	_ = lago.RegistryRoute.Register("nirmancampus_website.StaticRoute", lago.Route{
		Path:    "/nirman/static/{path...}",
		Handler: lago.NewDynamicView("nirmancampus_website.StaticView"),
	})

	_ = lago.RegistryRoute.Register("nirmancampus_website.PopupImageRoute", lago.Route{
		Path:    "/nirmancampus/popup-images/{id}/",
		Handler: lago.NewDynamicView("nirmancampus_website.PopupImageView"),
	})

	_ = lago.RegistryRoute.Register("nirmancampus_website.CoursesRoute", lago.Route{
		Path:    "/courses-offered/",
		Handler: lago.NewDynamicView("nirmancampus_website.CoursesView"),
	})

	// Overrides the root route (path "/", i.e. URL with no extra path segments).
	lago.RegistryRoute.Patch("base.HomeRoute", func(old lago.Route) lago.Route {
		old.Handler = lago.NewDynamicView("nirmancampus_website.HomeView")
		return old
	})
}
