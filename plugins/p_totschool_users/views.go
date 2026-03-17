package p_totschool_users

import (
	"net/http"

	"github.com/lariv-in/lago"
	"github.com/lariv-in/views"
)

// allowTotschoolAdminRolePatcher wraps the existing "users.role" middleware so that
// users with role totschool_admin are also allowed to access the view.
func allowTotschoolAdminRolePatcher(original views.Middleware) views.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if role, _ := r.Context().Value("$role").(string); role == "totschool_admin" {
				next.ServeHTTP(w, r)
				return
			}
			original(next).ServeHTTP(w, r)
		})
	}
}

// patchUsersRoleMiddleware applies the totschool_admin role patch to a view.
func patchUsersRoleMiddleware(v *views.View) *views.View {
	return v.PatchMiddleware("users.role", allowTotschoolAdminRolePatcher)
}

func init() {
	// Views that use "users.role" in p_users/views.go — allow totschool_admin to access them.
	userViewsWithRole := []string{
		"users.ListView",
		"users.DetailView",
		"users.CreateView",
		"users.UpdateView",
		"users.DeleteView",
		"users.ChangePasswordView",
		"users.SelectView",
		"users.MultiSelectView",
		"users.RoleSelectView",
		"users.RoleMultiSelectView",
		"users.RoleListView",
		"users.RoleDetailView",
		"users.RoleCreateView",
		"users.RoleUpdateView",
		"users.RoleDeleteView",
	}
	for _, name := range userViewsWithRole {
		lago.RegistryView.Patch(name, func(v *views.View) *views.View {
			return patchUsersRoleMiddleware(v)
		})
	}
}
