package p_nirmancampus_website

import (
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
)

func init() {
	lago.RegistryView.Register("nirmancampus_website.HomeView",
		lago.GetPageView("nirmancampus_website.HomePage").
			WithMiddleware("users.optional_auth", p_users.OptionalAuthMiddleware))

	lago.RegistryView.Register("nirmancampus_website.ProgramsView",
		lago.GetPageView("nirmancampus_website.ProgramsPage").
			WithMiddleware("users.optional_auth", p_users.OptionalAuthMiddleware))

	lago.RegistryView.Register("nirmancampus_website.ContactView",
		lago.GetPageView("nirmancampus_website.ContactPage").
			WithMiddleware("users.optional_auth", p_users.OptionalAuthMiddleware))

	lago.RegistryView.Register("nirmancampus_website.PrivacyView",
		lago.GetPageView("nirmancampus_website.PrivacyPage").
			WithMiddleware("users.optional_auth", p_users.OptionalAuthMiddleware))

	lago.RegistryView.Register("nirmancampus_website.StudentZoneView",
		lago.GetPageView("nirmancampus_website.StudentZonePage").
			WithMiddleware("users.optional_auth", p_users.OptionalAuthMiddleware))
}
