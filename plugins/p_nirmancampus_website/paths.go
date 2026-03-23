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

	_ = lago.RegistryRoute.Register("nirmancampus_website.AboutUsRoute", lago.Route{
		Path:    "/about-us/",
		Handler: lago.NewDynamicView("nirmancampus_website.AboutUsView"),
	})

	_ = lago.RegistryRoute.Register("nirmancampus_website.PrivacyPolicyRoute", lago.Route{
		Path:    "/privacy-policy/",
		Handler: lago.NewDynamicView("nirmancampus_website.PrivacyPolicyView"),
	})

	_ = lago.RegistryRoute.Register("nirmancampus_website.MrscmtRoute", lago.Route{
		Path:    "/mrscmt/",
		Handler: lago.NewDynamicView("nirmancampus_website.MrscmtView"),
	})

	_ = lago.RegistryRoute.Register("nirmancampus_website.MrsptuadmcoRoute", lago.Route{
		Path:    "/mrsptuadmco/",
		Handler: lago.NewDynamicView("nirmancampus_website.MrsptuadmcoView"),
	})

	_ = lago.RegistryRoute.Register("nirmancampus_website.Oss2281Route", lago.Route{
		Path:    "/oss2281/",
		Handler: lago.NewDynamicView("nirmancampus_website.Oss2281View"),
	})

	// Overrides the root route (path "/", i.e. URL with no extra path segments).
	lago.RegistryRoute.Patch("base.HomeRoute", func(old lago.Route) lago.Route {
		old.Handler = lago.NewDynamicView("nirmancampus_website.HomeView")
		return old
	})
}
