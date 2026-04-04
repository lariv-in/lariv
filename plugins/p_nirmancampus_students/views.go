package p_nirmancampus_students

import (
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/registry"
	"github.com/lariv-in/lago/views"
)

// studentsAdminRoleLayer limits create/update/delete to the admin role;
// superusers are always allowed (see p_users.RoleAuthorizationLayer).
var studentsAdminRoleLayer = p_users.RoleAuthorizationLayer{Roles: []string{"admin"}}

func init() {
	lago.RegistryView.Register("students.ListView",
		lago.GetPageView("students.StudentTable").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("students.list", views.LayerList[Student]{
				Key: getters.Static("students"),
				QueryPatchers: views.QueryPatchers[Student]{
					registry.Pair[string, views.QueryPatcher[Student]]{Key: "students.preload", Value: views.QueryPatcherPreload[Student]{Field: "User"}},
					registry.Pair[string, views.QueryPatcher[Student]]{Key: "students.scope_by_role", Value: StudentScopeByRole},
				},
			}))

	lago.RegistryView.Register("students.DetailView",
		lago.GetPageView("students.StudentDetail").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("students.detail", views.LayerDetail[Student]{
				Key:          getters.Static("student"),
				PathParamKey: getters.Static("id"),
				QueryPatchers: views.QueryPatchers[Student]{
					registry.Pair[string, views.QueryPatcher[Student]]{Key: "students.preload_user", Value: views.QueryPatcherPreload[Student]{Field: "User"}},
					registry.Pair[string, views.QueryPatcher[Student]]{Key: "students.preload_photo", Value: views.QueryPatcherPreload[Student]{Field: "Photo"}},
					registry.Pair[string, views.QueryPatcher[Student]]{Key: "students.preload_documents", Value: views.QueryPatcherPreload[Student]{Field: "Documents"}},
					registry.Pair[string, views.QueryPatcher[Student]]{Key: "students.scope_by_role", Value: StudentScopeByRole},
				},
			}))

	lago.RegistryView.Register("students.CreateView",
		lago.GetPageView("students.StudentCreateForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("students.admin_role", studentsAdminRoleLayer).
			WithLayer("students.create", views.LayerCreate[Student]{
				SuccessURL: lago.RoutePath("students.DetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("$id"))}),
			}))

	lago.RegistryView.Register("students.UpdateView",
		lago.GetPageView("students.StudentUpdateForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("students.admin_role", studentsAdminRoleLayer).
			WithLayer("students.detail", views.LayerDetail[Student]{
				Key:          getters.Static("student"),
				PathParamKey: getters.Static("id"),
				QueryPatchers: views.QueryPatchers[Student]{
					registry.Pair[string, views.QueryPatcher[Student]]{Key: "students.preload_user", Value: views.QueryPatcherPreload[Student]{Field: "User"}},
					registry.Pair[string, views.QueryPatcher[Student]]{Key: "students.preload_photo", Value: views.QueryPatcherPreload[Student]{Field: "Photo"}},
					registry.Pair[string, views.QueryPatcher[Student]]{Key: "students.preload_documents", Value: views.QueryPatcherPreload[Student]{Field: "Documents"}},
					registry.Pair[string, views.QueryPatcher[Student]]{Key: "students.scope_by_role", Value: StudentScopeByRole},
				},
			}).
			WithLayer("students.update", views.LayerUpdate[Student]{
				Key:        getters.Static("student"),
				SuccessURL: lago.RoutePath("students.DetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("student.ID"))}),
				QueryPatchers: views.QueryPatchers[Student]{
					registry.Pair[string, views.QueryPatcher[Student]]{Key: "students.scope_by_role", Value: StudentScopeByRole},
				},
			}))

	lago.RegistryView.Register("students.DeleteView",
		lago.GetPageView("students.StudentDeleteForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("students.admin_role", studentsAdminRoleLayer).
			WithLayer("students.detail", views.LayerDetail[Student]{
				Key:          getters.Static("student"),
				PathParamKey: getters.Static("id"),
				QueryPatchers: views.QueryPatchers[Student]{
					registry.Pair[string, views.QueryPatcher[Student]]{Key: "students.preload_user", Value: views.QueryPatcherPreload[Student]{Field: "User"}},
					registry.Pair[string, views.QueryPatcher[Student]]{Key: "students.preload_photo", Value: views.QueryPatcherPreload[Student]{Field: "Photo"}},
					registry.Pair[string, views.QueryPatcher[Student]]{Key: "students.preload_documents", Value: views.QueryPatcherPreload[Student]{Field: "Documents"}},
					registry.Pair[string, views.QueryPatcher[Student]]{Key: "students.scope_by_role", Value: StudentScopeByRole},
				},
			}).
			WithLayer("students.delete", views.LayerDelete[Student]{
				Key:        getters.Static("student"),
				SuccessURL: lago.RoutePath("students.DefaultRoute", nil),
				QueryPatchers: views.QueryPatchers[Student]{
					registry.Pair[string, views.QueryPatcher[Student]]{Key: "students.scope_by_role", Value: StudentScopeByRole},
				},
			}))

	lago.RegistryView.Register("students.SelectView",
		lago.GetPageView("students.StudentSelectionTable").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("students.select", views.LayerList[Student]{
				Key: getters.Static("students"),
				QueryPatchers: views.QueryPatchers[Student]{
					registry.Pair[string, views.QueryPatcher[Student]]{Key: "students.preload", Value: views.QueryPatcherPreload[Student]{Field: "User"}},
					registry.Pair[string, views.QueryPatcher[Student]]{Key: "students.scope_by_role", Value: StudentScopeByRole},
				},
			}))
}
