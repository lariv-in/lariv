package p_nirmancampus_website

import (
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
)

func init() {
	lago.RegistryView.Register("nirmancampus_website.HomeView",
		lago.GetPageView("nirmancampus_website.HomePage").
			WithLayer("users.optional_auth", p_users.OptionalAuthLayer{}))

	lago.RegistryView.Register("nirmancampus_website.ProgramsView",
		lago.GetPageView("nirmancampus_website.ProgramsPage").
			WithLayer("users.optional_auth", p_users.OptionalAuthLayer{}))

	lago.RegistryView.Register("nirmancampus_website.ContactView",
		lago.GetPageView("nirmancampus_website.ContactPage").
			WithLayer("users.optional_auth", p_users.OptionalAuthLayer{}))

	lago.RegistryView.Register("nirmancampus_website.PrivacyView",
		lago.GetPageView("nirmancampus_website.PrivacyPage").
			WithLayer("users.optional_auth", p_users.OptionalAuthLayer{}))

	lago.RegistryView.Register("nirmancampus_website.StudentZoneView",
		lago.GetPageView("nirmancampus_website.StudentZonePage").
			WithLayer("users.optional_auth", p_users.OptionalAuthLayer{}))
}
