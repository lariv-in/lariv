package p_nirmancampus_courses

import (
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/registry"
	"github.com/lariv-in/lago/views"
)

// coursesAdminRoleMiddleware limits create/update/delete to the admin role;
// superusers are always allowed (see p_users.RoleAuthorizationMiddleware).
var coursesAdminRoleMiddleware = p_users.RoleAuthorizationMiddleware{Roles: []string{"admin"}}

func init() {
	// List view
	lago.RegistryView.Register("courses.ListView",
		lago.GetPageView("courses.CourseTable").
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware{}).
			WithMiddleware("courses.list", views.MiddlewareList[Course]{
				Key: getters.Static("courses"),
				QueryPatchers: views.QueryPatchers[Course]{
					{Key: "courses.scope_by_role", Value: CourseScopeByRole},
				},
			}))

	// Detail view
	lago.RegistryView.Register("courses.DetailView",
		lago.GetPageView("courses.CourseDetail").
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware{}).
			WithMiddleware("courses.detail", views.MiddlewareDetail[Course]{
				Key:          getters.Static("course"),
				PathParamKey: getters.Static("id"),
				QueryPatchers: views.QueryPatchers[Course]{
					{Key: "courses.scope_by_role", Value: CourseScopeByRole},
				},
			}))

	// Create view
	lago.RegistryView.Register("courses.CreateView",
		lago.GetPageView("courses.CourseCreateForm").
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware{}).
			WithMiddleware("courses.admin_role", coursesAdminRoleMiddleware).
			WithMiddleware("courses.create", views.MiddlewareCreate[Course]{
				SuccessURL: lago.RoutePath("courses.DetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("$id"))}),
			}))

	// Update view
	lago.RegistryView.Register("courses.UpdateView",
		lago.GetPageView("courses.CourseUpdateForm").
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware{}).
			WithMiddleware("courses.admin_role", coursesAdminRoleMiddleware).
			WithMiddleware("courses.detail", views.MiddlewareDetail[Course]{
				Key:          getters.Static("course"),
				PathParamKey: getters.Static("id"),
				QueryPatchers: views.QueryPatchers[Course]{
					{Key: "courses.scope_by_role", Value: CourseScopeByRole},
				},
			}).
			WithMiddleware("courses.update", views.MiddlewareUpdate[Course]{
				Key:        getters.Static("course"),
				SuccessURL: lago.RoutePath("courses.DetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("course.ID"))}),
				QueryPatchers: views.QueryPatchers[Course]{
					{Key: "courses.scope_by_role", Value: CourseScopeByRole},
				},
			}))

	// Delete view
	lago.RegistryView.Register("courses.DeleteView",
		lago.GetPageView("courses.CourseDeleteForm").
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware{}).
			WithMiddleware("courses.admin_role", coursesAdminRoleMiddleware).
			WithMiddleware("courses.detail", views.MiddlewareDetail[Course]{
				Key:          getters.Static("course"),
				PathParamKey: getters.Static("id"),
				QueryPatchers: views.QueryPatchers[Course]{
					{Key: "courses.scope_by_role", Value: CourseScopeByRole},
				},
			}).
			WithMiddleware("courses.delete", views.MiddlewareDelete[Course]{
				Key:        getters.Static("course"),
				SuccessURL: lago.RoutePath("courses.DefaultRoute", nil),
				QueryPatchers: views.QueryPatchers[Course]{
					{Key: "courses.scope_by_role", Value: CourseScopeByRole},
				},
			}))

	// Selection views
	lago.RegistryView.Register("courses.SelectView",
		lago.GetPageView("courses.CourseSelectionTable").
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware{}).
			WithMiddleware("courses.select", views.MiddlewareList[Course]{
				Key: getters.Static("courses"),
				QueryPatchers: views.QueryPatchers[Course]{
					{Key: "courses.scope_by_role", Value: CourseScopeByRole},
				},
			}))

	lago.RegistryView.Register("courses.MultiSelectView",
		lago.GetPageView("courses.CourseMultiSelectionTable").
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware{}).
			WithMiddleware("courses.multiselect", views.MiddlewareList[Course]{
				Key: getters.Static("courses"),
				QueryPatchers: views.QueryPatchers[Course]{
					registry.Pair[string, views.QueryPatcher[Course]]{Key: "courses.scope_by_role", Value: CourseScopeByRole},
					registry.Pair[string, views.QueryPatcher[Course]]{Key: "courses.multiselect_pool_course_ids", Value: QueryPatcherMultiSelectPoolCourseIDs},
				},
			}))
}
