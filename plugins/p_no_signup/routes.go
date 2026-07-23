package p_no_signup

import (
	"net/http"

	"github.com/lariv-in/lariv"
	"github.com/lariv-in/lariv/plugins/p_users"
	"github.com/lariv-in/lariv/registry"
)

func pluginRoutes() lariv.PluginFeatures[lariv.Route] {
	return lariv.PluginFeatures[lariv.Route]{
		Patches: []registry.Pair[string, func(lariv.Route) lariv.Route]{
			{
				Key: "p_users.SignupRoute",
				Value: func(route lariv.Route) lariv.Route {
					route.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						http.Redirect(w, r, p_users.AppUrl+"login/", http.StatusSeeOther)
					})
					return route
				},
			},
		},
	}
}
