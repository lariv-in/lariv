package p_nirmancampus_website

import (
	"github.com/lariv-in/lago/lago"
)

func init() {
	_ = lago.RegistryRoute.Register("nirmancampus_website.StaticRoute", lago.Route{
		Path:    "/nirman/static/{path...}",
		Handler: lago.NewDynamicView("nirmancampus_website.StaticView"),
	})

	_ = lago.RegistryRoute.Register("nirmancampus_website.ContactRoute", lago.Route{
		Path:    "/contact-us/",
		Handler: lago.NewDynamicView("nirmancampus_website.ContactView"),
	})

	_ = lago.RegistryRoute.Register("nirmancampus_website.PrivacyRoute", lago.Route{
		Path:    "/privacy-policy/",
		Handler: lago.NewDynamicView("nirmancampus_website.PrivacyView"),
	})

	_ = lago.RegistryRoute.Register("nirmancampus_website.CoursesRoute", lago.Route{
		Path:    "/courses-offered/",
		Handler: lago.NewDynamicView("nirmancampus_website.CoursesView"),
	})

	_ = lago.RegistryRoute.Register("nirmancampus_website.ProgramsRoute", lago.Route{
		Path:    "/programs-offered/",
		Handler: lago.NewDynamicView("nirmancampus_website.ProgramsView"),
	})

	_ = lago.RegistryRoute.Register("nirmancampus_website.StudentZoneRoute", lago.Route{
		Path:    "/students-zone/",
		Handler: lago.NewDynamicView("nirmancampus_website.StudentZoneView"),
	})

	// Override the root route to serve the Nirman Campus home page.
	lago.RegistryRoute.Patch("base.HomeRoute", func(old lago.Route) lago.Route {
		old.Handler = lago.NewDynamicView("nirmancampus_website.HomeView")
		return old
	})
}
