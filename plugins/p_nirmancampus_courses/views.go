package p_nirmancampus_courses

import (
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/registry"
	"github.com/lariv-in/lago/views"
)

// coursesAdminRoleLayer limits create/update/delete to the admin role;
// superusers are always allowed (see p_users.RoleAuthorizationLayer).
var coursesAdminRoleLayer = p_users.RoleAuthorizationLayer{Roles: []string{"admin"}}

func init() {
	// List view
	lago.RegistryView.Register("courses.ListView",
		lago.GetPageView("courses.CourseTable").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("courses.list", views.LayerList[Course]{
				Key: getters.Static("courses"),
				QueryPatchers: views.QueryPatchers[Course]{
					{Key: "courses.scope_by_role", Value: CourseScopeByRole},
				},
			}))

	// Detail view
	lago.RegistryView.Register("courses.DetailView",
		lago.GetPageView("courses.CourseDetail").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("courses.detail", views.LayerDetail[Course]{
				Key:          getters.Static("course"),
				PathParamKey: getters.Static("id"),
				QueryPatchers: views.QueryPatchers[Course]{
					{Key: "courses.scope_by_role", Value: CourseScopeByRole},
				},
			}))

	// Create view
	lago.RegistryView.Register("courses.CreateView",
		lago.GetPageView("courses.CourseCreateForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("courses.admin_role", coursesAdminRoleLayer).
			WithLayer("courses.create", views.LayerCreate[Course]{
				SuccessURL: lago.RoutePath("courses.DetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("$id"))}),
			}))

	// Update view
	lago.RegistryView.Register("courses.UpdateView",
		lago.GetPageView("courses.CourseUpdateForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("courses.admin_role", coursesAdminRoleLayer).
			WithLayer("courses.detail", views.LayerDetail[Course]{
				Key:          getters.Static("course"),
				PathParamKey: getters.Static("id"),
				QueryPatchers: views.QueryPatchers[Course]{
					{Key: "courses.scope_by_role", Value: CourseScopeByRole},
				},
			}).
			WithLayer("courses.update", views.LayerUpdate[Course]{
				Key:        getters.Static("course"),
				SuccessURL: lago.RoutePath("courses.DetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("course.ID"))}),
				QueryPatchers: views.QueryPatchers[Course]{
					{Key: "courses.scope_by_role", Value: CourseScopeByRole},
				},
			}))

	// Delete view
	lago.RegistryView.Register("courses.DeleteView",
		lago.GetPageView("courses.CourseDeleteForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("courses.admin_role", coursesAdminRoleLayer).
			WithLayer("courses.detail", views.LayerDetail[Course]{
				Key:          getters.Static("course"),
				PathParamKey: getters.Static("id"),
				QueryPatchers: views.QueryPatchers[Course]{
					{Key: "courses.scope_by_role", Value: CourseScopeByRole},
				},
			}).
			WithLayer("courses.delete", views.LayerDelete[Course]{
				Key:        getters.Static("course"),
				SuccessURL: lago.RoutePath("courses.DefaultRoute", nil),
				QueryPatchers: views.QueryPatchers[Course]{
					{Key: "courses.scope_by_role", Value: CourseScopeByRole},
				},
			}))

	// Selection views
	lago.RegistryView.Register("courses.SelectView",
		lago.GetPageView("courses.CourseSelectionTable").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("courses.select", views.LayerList[Course]{
				Key: getters.Static("courses"),
				QueryPatchers: views.QueryPatchers[Course]{
					{Key: "courses.scope_by_role", Value: CourseScopeByRole},
				},
			}))

	lago.RegistryView.Register("courses.MultiSelectView",
		lago.GetPageView("courses.CourseMultiSelectionTable").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("courses.multiselect", views.LayerList[Course]{
				Key: getters.Static("courses"),
				QueryPatchers: views.QueryPatchers[Course]{
					registry.Pair[string, views.QueryPatcher[Course]]{Key: "courses.scope_by_role", Value: CourseScopeByRole},
					registry.Pair[string, views.QueryPatcher[Course]]{Key: "courses.multiselect_pool_course_ids", Value: QueryPatcherMultiSelectPoolCourseIDs},
				},
			}))
}
