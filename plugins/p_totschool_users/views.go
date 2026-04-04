package p_totschool_users

import (
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/views"

	"github.com/lariv-in/lago/plugins/p_users"
)

// userViewsWithRoleLayer are the users plugin views that use "users.role" layer.
var userViewsWithRoleLayer = []string{
	"users.ListView", "users.DetailView", "users.CreateView", "users.UpdateView",
	"users.DeleteView", "users.ChangePasswordView", "users.SelectView",
	"users.RoleSelectView", "users.RoleListView", "users.RoleDetailView",
	"users.RoleCreateView", "users.RoleUpdateView", "users.RoleDeleteView",
}

// Patch "users.role" layer to allow totschool_admin in addition to existing roles.
func patcher(current views.Layer) views.Layer {
	return p_users.RoleAuthorizationLayer{Roles: []string{"", "totschool_admin"}}
}

func init() {
	for _, key := range userViewsWithRoleLayer {
		if v, ok := lago.RegistryView.Get(key); ok {
			v.PatchLayer("users.role", patcher)
		}
	}
}
