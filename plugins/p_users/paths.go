package p_users

import (
	"github.com/lariv-in/lariv"
	"github.com/lariv-in/lariv/registry"
)

const (
	// Trailing slashes so AppUrl+"{…}/" concatenation yields whole wildcard segments for ServeMux.
	AppUrl  = "/users/"
	RoleUrl = "/users/roles/"
	// Routes keyed by DB user ID live under …/u/{id}/… so literals like …/roles/… never match …/{id}/….
	UserIDRoutePrefix = AppUrl + "u/"
)

func pluginRoutes() lariv.PluginFeatures[lariv.Route] {
	return lariv.PluginFeatures[lariv.Route]{
		Patches: []registry.Pair[string, func(lariv.Route) lariv.Route]{
			{
				Key: "core.HomeRoute",
				Value: func(old lariv.Route) lariv.Route {
					old.Path = "/"
					old.Handler = lariv.NewDynamicView("core.HomeView")
					return old
				},
			},
		},
		Entries: []registry.Pair[string, lariv.Route]{
			{Key: "p_users.ListRoute", Value: lariv.Route{
				Path:    AppUrl,
				Handler: lariv.NewDynamicView("p_users.ListView"),
			}},
			{Key: "p_users.CreateRoute", Value: lariv.Route{
				Path:    AppUrl + "create/",
				Handler: lariv.NewDynamicView("p_users.CreateView"),
			}},
			{Key: "p_users.DetailRoute", Value: lariv.Route{
				Path:    UserIDRoutePrefix + "{id}/",
				Handler: lariv.NewDynamicView("p_users.DetailView"),
			}},
			{Key: "p_users.UpdateRoute", Value: lariv.Route{
				Path:    UserIDRoutePrefix + "{id}/edit/",
				Handler: lariv.NewDynamicView("p_users.UpdateView"),
			}},
			{Key: "p_users.SelfDetailRoute", Value: lariv.Route{
				Path:    AppUrl + "self/",
				Handler: lariv.NewDynamicView("p_users.SelfDetailView"),
			}},
			{Key: "p_users.SelfUpdateRoute", Value: lariv.Route{
				Path:    AppUrl + "self/edit/",
				Handler: lariv.NewDynamicView("p_users.SelfUpdateView"),
			}},
			{Key: "p_users.SelfChangePasswordRoute", Value: lariv.Route{
				Path:    AppUrl + "self/change-password/",
				Handler: lariv.NewDynamicView("p_users.SelfChangePasswordView"),
			}},
			{Key: "p_users.DeleteRoute", Value: lariv.Route{
				Path:    UserIDRoutePrefix + "{id}/delete/",
				Handler: lariv.NewDynamicView("p_users.DeleteView"),
			}},
			{Key: "p_users.ChangePasswordRoute", Value: lariv.Route{
				Path:    UserIDRoutePrefix + "{id}/change-password/",
				Handler: lariv.NewDynamicView("p_users.ChangePasswordView"),
			}},
			{Key: "p_users.SelectRoute", Value: lariv.Route{
				Path:    AppUrl + "select/",
				Handler: lariv.NewDynamicView("p_users.SelectView"),
			}},
			{Key: "p_users.RoleSelectRoute", Value: lariv.Route{
				Path:    RoleUrl + "select/",
				Handler: lariv.NewDynamicView("p_users.RoleSelectView"),
			}},
			{Key: "p_users.RoleListRoute", Value: lariv.Route{
				Path:    RoleUrl,
				Handler: lariv.NewDynamicView("p_users.RoleListView"),
			}},
			{Key: "p_users.RoleCreateRoute", Value: lariv.Route{
				Path:    RoleUrl + "create/",
				Handler: lariv.NewDynamicView("p_users.RoleCreateView"),
			}},
			{Key: "p_users.RoleDetailRoute", Value: lariv.Route{
				Path:    RoleUrl + "{id}/",
				Handler: lariv.NewDynamicView("p_users.RoleDetailView"),
			}},
			{Key: "p_users.RoleUpdateRoute", Value: lariv.Route{
				Path:    RoleUrl + "{id}/edit/",
				Handler: lariv.NewDynamicView("p_users.RoleUpdateView"),
			}},
			{Key: "p_users.RoleDeleteRoute", Value: lariv.Route{
				Path:    RoleUrl + "{id}/delete/",
				Handler: lariv.NewDynamicView("p_users.RoleDeleteView"),
			}},
			{Key: "p_users.LoginRoute", Value: lariv.Route{
				Path:    AppUrl + "login/",
				Handler: lariv.NewDynamicView("p_users.LoginView"),
			}},
			{Key: "p_users.SignupRoute", Value: lariv.Route{
				Path:    AppUrl + "signup/",
				Handler: lariv.NewDynamicView("p_users.SignupView"),
			}},
			{Key: "p_users.LoginSuccessRoute", Value: lariv.Route{
				Path:    AppUrl + "success/",
				Handler: lariv.NewDynamicView("p_users.LoginSuccessView"),
			}},
			{Key: "p_users.UnauthenticatedRoute", Value: lariv.Route{
				Path:    AppUrl + "unauthenticated/",
				Handler: lariv.NewDynamicView("p_users.UnauthenticatedView"),
			}},
			{Key: "p_users.LogoutRoute", Value: lariv.Route{
				Path:    AppUrl + "logout/",
				Handler: lariv.NewDynamicView("p_users.LogoutView"),
			}},
		},
	}
}
