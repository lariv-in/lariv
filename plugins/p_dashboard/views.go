package p_dashboard

import (
	"context"

	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/views"
)

func init() {
	lago.RegistryView.Register("dashboard.AppsView",
		lago.GetPageView("dashboard.AppsPage").WithLayer("users.auth", p_users.AuthenticationLayer{}))
	lago.RegistryView.Patch("users.LoginSuccessView", func(_ *views.View) *views.View {
		return lago.RedirectView(lago.RoutePath("dashboard.AppsPage", nil))
	})

	// base.HomeRoute uses view base.HomeView: send logged-in users to apps grid, others to login.
	lago.RegistryView.Patch("base.HomeView", func(_ *views.View) *views.View {
		return lago.GetPageView("dashboard.HomeRedirectStub").
			WithLayer("users.optional_auth", p_users.OptionalAuthLayer{}).
			WithLayer("dashboard.home_root_redirect", lago.RedirectLayer{URLGetter: func(ctx context.Context) (string, error) {
				if p_users.UserPresentInContext(ctx) {
					return lago.RoutePath("dashboard.AppsPage", nil)(ctx)
				}
				return lago.RoutePath("users.LoginRoute", nil)(ctx)
			}})
	})
}
