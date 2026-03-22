package p_nirmancampus_users

import (
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/views"

	"github.com/lariv-in/lago/p_users"
)

// userViewsWithRoleMiddleware are the users plugin views that use "users.role" middleware.
var userViewsWithRoleMiddleware = []string{
	"users.ListView", "users.DetailView", "users.CreateView", "users.UpdateView",
	"users.DeleteView", "users.ChangePasswordView", "users.SelectView",
	"users.RoleSelectView", "users.RoleListView", "users.RoleDetailView",
}

// Patch "users.role" middleware to allow nirmancampus_admin in addition to existing roles.
func patcher(current views.Middleware) views.Middleware {
	return p_users.RoleAuthorizationMiddleware([]string{"nirmancampus_admin"})
}

func init() {
	for _, key := range userViewsWithRoleMiddleware {
		if v, ok := lago.RegistryView.Get(key); ok {
			v.PatchMiddleware("users.role", patcher)
		}
	}
}
