package p_nirmancampus_website

import (
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
)

func init() {
	lago.RegistryView.Register("nirmancampus_website.AppLandingView",
		lago.GetPageView("nirmancampus_website.AppLandingPage").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("website.role", p_users.RoleAuthorizationLayer{Roles: []string{"admin"}}))
}
