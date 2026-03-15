package p_dashboard

import (
	"net/http"

	"github.com/lariv-in/lago"
	"github.com/lariv-in/p_users"
)


func init() {
	lago.RegistryView.Register("dashboard.AppsView", p_users.AuthMiddleware(lago.GetPageView("dashboard.AppsPage")))
	lago.RegistryView.Patch("users.LoginSuccessView", func(_ http.Handler) http.Handler {
		return lago.NewRedirectView("dashboard.AppsPage")
	})
}
