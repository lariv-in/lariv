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
	_ = lago.RegistryRoute.Register("users.AllUsersRoute", lago.Route{
		Path:    AppUrl,
		Handler: lago.NewDynamicView("users.AllUsersView"),
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
