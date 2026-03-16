package p_dashboard

import (
	"github.com/lariv-in/lago"
	"github.com/lariv-in/p_users"
	"github.com/lariv-in/views"
)


func init() {
	lago.RegistryView.Register("dashboard.AppsView",
		lago.GetPageView("dashboard.AppsPage").WithMiddleware("users.auth", p_users.AuthenticationMiddleware))
	lago.RegistryView.Patch("users.LoginSuccessView", func(_ *views.View) *views.View {
		return lago.NewRedirectView("dashboard.AppsPage")
	})
}
