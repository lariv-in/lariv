package p_nirmancampus_website

import (
	"github.com/lariv-in/lago/lago"
)

func init() {
	_ = lago.RegistryRoute.Register("nirmancampus_website.PopupImageRoute", lago.Route{
		Path:    "/nirmancampus/popup-images/{id}/",
		Handler: lago.NewDynamicView("nirmancampus_website.PopupImageView"),
	})

	// Overrides the root route (path "/", i.e. URL with no extra path segments).
	lago.RegistryRoute.Patch("base.HomeRoute", func(old lago.Route) lago.Route {
		old.Handler = lago.NewDynamicView("nirmancampus_website.HomeView")
		return old
	})
}
