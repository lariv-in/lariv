package p_dashboard

import (
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/p_users"
	"github.com/lariv-in/lago/views"
)


func init() {
	lago.RegistryView.Register("dashboard.AppsView",
		lago.GetPageView("dashboard.AppsPage").WithMiddleware("users.auth", p_users.AuthenticationMiddleware))
	lago.RegistryView.Patch("users.LoginSuccessView", func(_ *views.View) *views.View {
		return lago.NewRedirectView("dashboard.AppsPage")
	})
}
