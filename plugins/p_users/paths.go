package p_users

import (
	"github.com/lariv-in/lago/lago"
)

func registerRoutes() {
	err := lago.RegistryRoute.Register("base.HomeRoute", lago.Route{
		Path:    "/",
		Handler: lago.NewDynamicView("base.HomeView"),
	})
	if err != nil {
		lago.RegistryRoute.Patch("base.HomePage", func(oldRoute lago.Route) lago.Route {
			oldRoute.Handler = lago.NewDynamicView("base.HomeView")
			return oldRoute
		})
	}
	_ = lago.RegistryRoute.Register("users.ListRoute", lago.Route{
		Path:    AppUrl,
		Handler: lago.NewDynamicView("users.ListView"),
	})

	_ = lago.RegistryRoute.Register("users.CreateRoute", lago.Route{
		Path:    AppUrl + "create/",
		Handler: lago.NewDynamicView("users.CreateView"),
	})

	_ = lago.RegistryRoute.Register("users.DetailRoute", lago.Route{
		Path:    AppUrl + "{id}/",
		Handler: lago.NewDynamicView("users.DetailView"),
	})

	_ = lago.RegistryRoute.Register("users.UpdateRoute", lago.Route{
		Path:    AppUrl + "{id}/edit/",
		Handler: lago.NewDynamicView("users.UpdateView"),
	})

	_ = lago.RegistryRoute.Register("users.SelfDetailRoute", lago.Route{
		Path:    AppUrl + "self/",
		Handler: lago.NewDynamicView("users.SelfDetailView"),
	})

	_ = lago.RegistryRoute.Register("users.SelfUpdateRoute", lago.Route{
		Path:    AppUrl + "self/edit/",
		Handler: lago.NewDynamicView("users.SelfUpdateView"),
	})

	_ = lago.RegistryRoute.Register("users.SelfChangePasswordRoute", lago.Route{
		Path:    AppUrl + "self/change-password/",
		Handler: lago.NewDynamicView("users.SelfChangePasswordView"),
	})

	_ = lago.RegistryRoute.Register("users.DeleteRoute", lago.Route{
		Path:    AppUrl + "{id}/delete/",
		Handler: lago.NewDynamicView("users.DeleteView"),
	})

	_ = lago.RegistryRoute.Register("users.ChangePasswordRoute", lago.Route{
		Path:    AppUrl + "{id}/change-password/",
		Handler: lago.NewDynamicView("users.ChangePasswordView"),
	})

	_ = lago.RegistryRoute.Register("users.SelectRoute", lago.Route{
		Path:    AppUrl + "select/",
		Handler: lago.NewDynamicView("users.SelectView"),
	})

	_ = lago.RegistryRoute.Register("users.RoleSelectRoute", lago.Route{
		Path:    RoleUrl + "select/",
		Handler: lago.NewDynamicView("users.RoleSelectView"),
	})

	_ = lago.RegistryRoute.Register("users.RoleListRoute", lago.Route{
		Path:    RoleUrl,
		Handler: lago.NewDynamicView("users.RoleListView"),
	})

	_ = lago.RegistryRoute.Register("users.RoleCreateRoute", lago.Route{
		Path:    RoleUrl + "create/",
		Handler: lago.NewDynamicView("users.RoleCreateView"),
	})

	_ = lago.RegistryRoute.Register("users.RoleDetailRoute", lago.Route{
		Path:    RoleUrl + "{id}/",
		Handler: lago.NewDynamicView("users.RoleDetailView"),
	})

	_ = lago.RegistryRoute.Register("users.RoleUpdateRoute", lago.Route{
		Path:    RoleUrl + "{id}/edit/",
		Handler: lago.NewDynamicView("users.RoleUpdateView"),
	})

	_ = lago.RegistryRoute.Register("users.RoleDeleteRoute", lago.Route{
		Path:    RoleUrl + "{id}/delete/",
		Handler: lago.NewDynamicView("users.RoleDeleteView"),
	})

	_ = lago.RegistryRoute.Register("users.LoginRoute", lago.Route{
		Path:    AppUrl + "login/",
		Handler: lago.NewDynamicView("users.LoginView"),
	})

	_ = lago.RegistryRoute.Register("users.SignupRoute", lago.Route{
		Path:    AppUrl + "signup/",
		Handler: lago.NewDynamicView("users.SignupView"),
	})

	_ = lago.RegistryRoute.Register("users.LoginSuccessRoute", lago.Route{
		Path:    AppUrl + "success/",
		Handler: lago.NewDynamicView("users.LoginSuccessView"),
	})

	_ = lago.RegistryRoute.Register("users.UnauthenticatedRoute", lago.Route{
		Path:    AppUrl + "unauthenticated/",
		Handler: lago.NewDynamicView("users.UnauthenticatedView"),
	})

	_ = lago.RegistryRoute.Register("users.LogoutRoute", lago.Route{
		Path:    AppUrl + "logout/",
		Handler: lago.NewDynamicView("users.LogoutView"),
	})
}

func init() {
	registerRoutes()
}
