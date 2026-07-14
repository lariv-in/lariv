package p_dashboard

import (
	"context"

	"github.com/lariv-in/lago"
	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/registry"
	"github.com/lariv-in/lago/views"
)

func pluginViews() lago.PluginFeatures[*views.View] {
	return lago.PluginFeatures[*views.View]{
		Entries: []registry.Pair[string, *views.View]{
			{Key: "dashboard.AppsView", Value: lago.GetPageView("dashboard.AppsPage").WithLayer("p_users.auth", p_users.AuthenticationLayer{})},
		},
		Patches: []registry.Pair[string, func(*views.View) *views.View]{
			{Key: "p_users.LoginSuccessView", Value: func(_ *views.View) *views.View {
				return lago.RedirectView(lago.RoutePath("dashboard.AppsPage", nil))
			}},
			// core.HomeView: core.HomeRoute renders this; dashboard sends logged-in users to apps, others to login.
			{Key: "core.HomeView", Value: func(_ *views.View) *views.View {
				return lago.GetPageView("dashboard.HomeRedirectStub").
					WithLayer("p_users.optional_auth", p_users.OptionalAuthLayer{}).
					WithLayer("dashboard.home_root_redirect", lago.RedirectLayer{URLGetter: func(ctx context.Context) (string, error) {
						if p_users.UserPresentInContext(ctx) {
							return lago.RoutePath("dashboard.AppsPage", nil)(ctx)
						}
						return lago.RoutePath("p_users.LoginRoute", nil)(ctx)
					}})
			}},
		},
	}
}
