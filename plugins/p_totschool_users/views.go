package p_totschool_users

import (
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/views"

	"github.com/lariv-in/lago/plugins/p_users"
)

// userViewsWithRoleMiddleware are the users plugin views that use "users.role" middleware.
var userViewsWithRoleMiddleware = []string{
	"users.ListView", "users.DetailView", "users.CreateView", "users.UpdateView",
	"users.DeleteView", "users.ChangePasswordView", "users.SelectView",
	"users.RoleSelectView", "users.RoleListView", "users.RoleDetailView",
	"users.RoleCreateView", "users.RoleUpdateView", "users.RoleDeleteView",
}

// Patch "users.role" middleware to allow totschool_admin in addition to existing roles.
func patcher(current views.Middleware) views.Middleware {
	return p_users.RoleAuthorizationMiddleware([]string{"", "totschool_admin"})
}

func init() {
	for _, key := range userViewsWithRoleMiddleware {
		if v, ok := lago.RegistryView.Get(key); ok {
			v.PatchMiddleware("users.role", patcher)
		}
	}
}
