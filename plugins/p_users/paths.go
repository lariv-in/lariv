package p_users

import (
	"github.com/lariv-in/lago"
	"github.com/lariv-in/lago/registry"
)

const (
	// Trailing slashes so AppUrl+"{…}/" concatenation yields whole wildcard segments for ServeMux.
	AppUrl  = "/users/"
	RoleUrl = "/users/roles/"
	// Routes keyed by DB user ID live under …/u/{id}/… so literals like …/roles/… never match …/{id}/….
	UserIDRoutePrefix = AppUrl + "u/"
)

func pluginRoutes() lago.PluginFeatures[lago.Route] {
	return lago.PluginFeatures[lago.Route]{
		Patches: []registry.Pair[string, func(lago.Route) lago.Route]{
			{
				Key: "core.HomeRoute",
				Value: func(old lago.Route) lago.Route {
					old.Path = "/"
					old.Handler = lago.NewDynamicView("core.HomeView")
					return old
				},
			},
		},
		Entries: []registry.Pair[string, lago.Route]{
			{Key: "p_users.ListRoute", Value: lago.Route{
				Path:    AppUrl,
				Handler: lago.NewDynamicView("p_users.ListView"),
			}},
			{Key: "p_users.CreateRoute", Value: lago.Route{
				Path:    AppUrl + "create/",
				Handler: lago.NewDynamicView("p_users.CreateView"),
			}},
			{Key: "p_users.DetailRoute", Value: lago.Route{
				Path:    UserIDRoutePrefix + "{id}/",
				Handler: lago.NewDynamicView("p_users.DetailView"),
			}},
			{Key: "p_users.UpdateRoute", Value: lago.Route{
				Path:    UserIDRoutePrefix + "{id}/edit/",
				Handler: lago.NewDynamicView("p_users.UpdateView"),
			}},
			{Key: "p_users.SelfDetailRoute", Value: lago.Route{
				Path:    AppUrl + "self/",
				Handler: lago.NewDynamicView("p_users.SelfDetailView"),
			}},
			{Key: "p_users.SelfUpdateRoute", Value: lago.Route{
				Path:    AppUrl + "self/edit/",
				Handler: lago.NewDynamicView("p_users.SelfUpdateView"),
			}},
			{Key: "p_users.SelfChangePasswordRoute", Value: lago.Route{
				Path:    AppUrl + "self/change-password/",
				Handler: lago.NewDynamicView("p_users.SelfChangePasswordView"),
			}},
			{Key: "p_users.DeleteRoute", Value: lago.Route{
				Path:    UserIDRoutePrefix + "{id}/delete/",
				Handler: lago.NewDynamicView("p_users.DeleteView"),
			}},
			{Key: "p_users.ChangePasswordRoute", Value: lago.Route{
				Path:    UserIDRoutePrefix + "{id}/change-password/",
				Handler: lago.NewDynamicView("p_users.ChangePasswordView"),
			}},
			{Key: "p_users.SelectRoute", Value: lago.Route{
				Path:    AppUrl + "select/",
				Handler: lago.NewDynamicView("p_users.SelectView"),
			}},
			{Key: "p_users.RoleSelectRoute", Value: lago.Route{
				Path:    RoleUrl + "select/",
				Handler: lago.NewDynamicView("p_users.RoleSelectView"),
			}},
			{Key: "p_users.RoleListRoute", Value: lago.Route{
				Path:    RoleUrl,
				Handler: lago.NewDynamicView("p_users.RoleListView"),
			}},
			{Key: "p_users.RoleCreateRoute", Value: lago.Route{
				Path:    RoleUrl + "create/",
				Handler: lago.NewDynamicView("p_users.RoleCreateView"),
			}},
			{Key: "p_users.RoleDetailRoute", Value: lago.Route{
				Path:    RoleUrl + "{id}/",
				Handler: lago.NewDynamicView("p_users.RoleDetailView"),
			}},
			{Key: "p_users.RoleUpdateRoute", Value: lago.Route{
				Path:    RoleUrl + "{id}/edit/",
				Handler: lago.NewDynamicView("p_users.RoleUpdateView"),
			}},
			{Key: "p_users.RoleDeleteRoute", Value: lago.Route{
				Path:    RoleUrl + "{id}/delete/",
				Handler: lago.NewDynamicView("p_users.RoleDeleteView"),
			}},
			{Key: "p_users.LoginRoute", Value: lago.Route{
				Path:    AppUrl + "login/",
				Handler: lago.NewDynamicView("p_users.LoginView"),
			}},
			{Key: "p_users.SignupRoute", Value: lago.Route{
				Path:    AppUrl + "signup/",
				Handler: lago.NewDynamicView("p_users.SignupView"),
			}},
			{Key: "p_users.LoginSuccessRoute", Value: lago.Route{
				Path:    AppUrl + "success/",
				Handler: lago.NewDynamicView("p_users.LoginSuccessView"),
			}},
			{Key: "p_users.UnauthenticatedRoute", Value: lago.Route{
				Path:    AppUrl + "unauthenticated/",
				Handler: lago.NewDynamicView("p_users.UnauthenticatedView"),
			}},
			{Key: "p_users.LogoutRoute", Value: lago.Route{
				Path:    AppUrl + "logout/",
				Handler: lago.NewDynamicView("p_users.LogoutView"),
			}},
		},
	}
}
