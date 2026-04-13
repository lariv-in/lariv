package p_nirmancampus_website

import "github.com/lariv-in/lago/lago"

func init() {
	_ = lago.RegistryRoute.Register("nirmancampus_website.AppLandingRoute", lago.Route{
		Path:    AppUrl,
		Handler: lago.NewDynamicView("nirmancampus_website.AppLandingView"),
	})
}
