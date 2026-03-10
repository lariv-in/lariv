package p_users

import (
	"log"

	"github.com/lariv-in/lago"
)

func registerRoutes() {
	err := lago.RegistryRoute.Register("base.HomeRoute", lago.Route{
		Path:    "/",
		Handler: lago.NewDynamicView("base.HomeView"),
	})
	if err != nil {
		err2 := lago.RegistryRoute.Patch("base.HomePage", func(oldRoute lago.Route) lago.Route {
			oldRoute.Handler = lago.NewDynamicView("base.HomeView")
			return oldRoute
		})
		if err2 != nil {
			log.Panicf("Can't register, Can't patch, something wierd is going on.\nRegister Error: %e\nPatch Error: %e", err, err2)
		}
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

	_ = lago.RegistryRoute.Register("users.MultiSelectRoute", lago.Route{
		Path:    AppUrl + "multi-select/",
		Handler: lago.NewDynamicView("users.MultiSelectView"),
	})

	_ = lago.RegistryRoute.Register("users.RoleSelectRoute", lago.Route{
		Path:    AppUrl + "roles/select/",
		Handler: lago.NewDynamicView("users.RoleSelectView"),
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
