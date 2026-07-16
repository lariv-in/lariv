package p_dashboard

import (
	"context"

	"github.com/lariv-in/lariv"
	"github.com/lariv-in/lariv/plugins/p_users"
	"github.com/lariv-in/lariv/registry"
	"github.com/lariv-in/lariv/views"
)

func pluginViews() lariv.PluginFeatures[*views.View] {
	return lariv.PluginFeatures[*views.View]{
		Entries: []registry.Pair[string, *views.View]{
			{Key: "dashboard.AppsView", Value: lariv.GetPageView("dashboard.AppsPage").WithLayer("p_users.auth", p_users.AuthenticationLayer{})},
		},
		Patches: []registry.Pair[string, func(*views.View) *views.View]{
			{Key: "p_users.LoginSuccessView", Value: func(_ *views.View) *views.View {
				return lariv.RedirectView(lariv.RoutePath("dashboard.AppsPage", nil))
			}},
			// core.HomeView: core.HomeRoute renders this; dashboard sends logged-in users to apps, others to login.
			{Key: "core.HomeView", Value: func(_ *views.View) *views.View {
				return lariv.GetPageView("dashboard.HomeRedirectStub").
					WithLayer("p_users.optional_auth", p_users.OptionalAuthLayer{}).
					WithLayer("dashboard.home_root_redirect", lariv.RedirectLayer{URLGetter: func(ctx context.Context) (string, error) {
						if p_users.UserPresentInContext(ctx) {
							return lariv.RoutePath("dashboard.AppsPage", nil)(ctx)
						}
						return lariv.RoutePath("p_users.LoginRoute", nil)(ctx)
					}})
			}},
		},
	}
}
