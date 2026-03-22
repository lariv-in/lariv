package p_dashboard

import (
	"github.com/lariv-in/lago/lago"
)

func init() {
	lago.RegistryRoute.Register("dashboard.AppsPage", lago.Route{
		Path:    "/apps/",
		Handler: lago.NewDynamicView("dashboard.AppsView"),
	})
}
