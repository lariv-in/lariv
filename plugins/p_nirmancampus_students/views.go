package p_nirmancampus_students

import (
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/registry"
	"github.com/lariv-in/lago/views"
)

// studentsAdminRoleMiddleware limits create/update/delete to the admin role;
// superusers are always allowed (see p_users.RoleAuthorizationMiddleware).
var studentsAdminRoleMiddleware = p_users.RoleAuthorizationMiddleware{Roles: []string{"admin"}}

func init() {
	lago.RegistryView.Register("students.ListView",
		lago.GetPageView("students.StudentTable").
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware{}).
			WithMiddleware("students.list", views.MiddlewareList[Student]{
				Key: getters.Static("students"),
				QueryPatchers: views.QueryPatchers[Student]{
					registry.Pair[string, views.QueryPatcher[Student]]{Key: "students.preload", Value: views.QueryPatcherPreload[Student]{Field: "User"}},
					registry.Pair[string, views.QueryPatcher[Student]]{Key: "students.scope_by_role", Value: StudentScopeByRole},
				},
			}))

	lago.RegistryView.Register("students.DetailView",
		lago.GetPageView("students.StudentDetail").
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware{}).
			WithMiddleware("students.detail", views.MiddlewareDetail[Student]{
				Key:          getters.Static("student"),
				PathParamKey: getters.Static("id"),
				QueryPatchers: views.QueryPatchers[Student]{
					registry.Pair[string, views.QueryPatcher[Student]]{Key: "students.preload_user", Value: views.QueryPatcherPreload[Student]{Field: "User"}},
					registry.Pair[string, views.QueryPatcher[Student]]{Key: "students.preload_assets", Value: views.QueryPatcherPreload[Student]{Field: "Assets"}},
					registry.Pair[string, views.QueryPatcher[Student]]{Key: "students.scope_by_role", Value: StudentScopeByRole},
				},
			}))

	lago.RegistryView.Register("students.CreateView",
		lago.GetPageView("students.StudentCreateForm").
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware{}).
			WithMiddleware("students.admin_role", studentsAdminRoleMiddleware).
			WithMiddleware("students.create", views.MiddlewareCreate[Student]{
				SuccessURL: lago.RoutePath("students.DetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("$id"))}),
			}))

	lago.RegistryView.Register("students.UpdateView",
		lago.GetPageView("students.StudentUpdateForm").
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware{}).
			WithMiddleware("students.admin_role", studentsAdminRoleMiddleware).
			WithMiddleware("students.detail", views.MiddlewareDetail[Student]{
				Key:          getters.Static("student"),
				PathParamKey: getters.Static("id"),
				QueryPatchers: views.QueryPatchers[Student]{
					registry.Pair[string, views.QueryPatcher[Student]]{Key: "students.preload_user", Value: views.QueryPatcherPreload[Student]{Field: "User"}},
					registry.Pair[string, views.QueryPatcher[Student]]{Key: "students.preload_assets", Value: views.QueryPatcherPreload[Student]{Field: "Assets"}},
					registry.Pair[string, views.QueryPatcher[Student]]{Key: "students.scope_by_role", Value: StudentScopeByRole},
				},
			}).
			WithMiddleware("students.update", views.MiddlewareUpdate[Student]{
				Key:        getters.Static("student"),
				SuccessURL: lago.RoutePath("students.DetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("student.ID"))}),
				QueryPatchers: views.QueryPatchers[Student]{
					registry.Pair[string, views.QueryPatcher[Student]]{Key: "students.scope_by_role", Value: StudentScopeByRole},
				},
			}))

	lago.RegistryView.Register("students.DeleteView",
		lago.GetPageView("students.StudentDeleteForm").
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware{}).
			WithMiddleware("students.admin_role", studentsAdminRoleMiddleware).
			WithMiddleware("students.detail", views.MiddlewareDetail[Student]{
				Key:          getters.Static("student"),
				PathParamKey: getters.Static("id"),
				QueryPatchers: views.QueryPatchers[Student]{
					registry.Pair[string, views.QueryPatcher[Student]]{Key: "students.preload_user", Value: views.QueryPatcherPreload[Student]{Field: "User"}},
					registry.Pair[string, views.QueryPatcher[Student]]{Key: "students.preload_assets", Value: views.QueryPatcherPreload[Student]{Field: "Assets"}},
					registry.Pair[string, views.QueryPatcher[Student]]{Key: "students.scope_by_role", Value: StudentScopeByRole},
				},
			}).
			WithMiddleware("students.delete", views.MiddlewareDelete[Student]{
				Key:        getters.Static("student"),
				SuccessURL: lago.RoutePath("students.DefaultRoute", nil),
				QueryPatchers: views.QueryPatchers[Student]{
					registry.Pair[string, views.QueryPatcher[Student]]{Key: "students.scope_by_role", Value: StudentScopeByRole},
				},
			}))

	lago.RegistryView.Register("students.SelectView",
		lago.GetPageView("students.StudentSelectionTable").
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware{}).
			WithMiddleware("students.select", views.MiddlewareList[Student]{
				Key: getters.Static("students"),
				QueryPatchers: views.QueryPatchers[Student]{
					registry.Pair[string, views.QueryPatcher[Student]]{Key: "students.preload", Value: views.QueryPatcherPreload[Student]{Field: "User"}},
					registry.Pair[string, views.QueryPatcher[Student]]{Key: "students.scope_by_role", Value: StudentScopeByRole},
				},
			}))
}
