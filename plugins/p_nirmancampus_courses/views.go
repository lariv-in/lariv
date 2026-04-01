package p_nirmancampus_courses

import (
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/views"
)

// coursesAdminRoleMiddleware limits create/update/delete to the admin role;
// superusers are always allowed (see p_users.RoleAuthorizationMiddleware).
var coursesAdminRoleMiddleware = p_users.RoleAuthorizationMiddleware([]string{"admin"})

func init() {
	// List view
	lago.RegistryView.Register("courses.ListView",
		views.ListView[Course]("courses")(
			lago.GetPageView("courses.CourseTable")).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithQueryPatcher("courses.scope_by_role", CourseScopeByRole))

	// Detail view
	lago.RegistryView.Register("courses.DetailView",
		views.DetailView[Course]("course", "id")(
			lago.GetPageView("courses.CourseDetail")).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithQueryPatcher("courses.scope_by_role", CourseScopeByRole))

	// Create view
	lago.RegistryView.Register("courses.CreateView",
		views.CreateView[Course](lago.GetterRoutePath("courses.DetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("$id"))}))(
			lago.GetPageView("courses.CourseCreateForm")).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithMiddleware("courses.admin_role", coursesAdminRoleMiddleware))

	// Update view
	lago.RegistryView.Register("courses.UpdateView",
		views.DetailView[Course]("course", "id")(
			views.UpdateView[Course]("id", lago.GetterRoutePath("courses.DetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("$id"))}))(
				lago.GetPageView("courses.CourseUpdateForm"))).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithMiddleware("courses.admin_role", coursesAdminRoleMiddleware).
			WithQueryPatcher("courses.scope_by_role", CourseScopeByRole))

	// Delete view
	lago.RegistryView.Register("courses.DeleteView",
		views.DetailView[Course]("course", "id")(
			views.DeleteView[Course]("id", lago.GetterRoutePath("courses.DefaultRoute", nil))(
				lago.GetPageView("courses.CourseDeleteForm"))).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithMiddleware("courses.admin_role", coursesAdminRoleMiddleware).
			WithQueryPatcher("courses.scope_by_role", CourseScopeByRole))

	// Selection views
	lago.RegistryView.Register("courses.SelectView",
		views.ListView[Course]("courses")(
			lago.GetPageView("courses.CourseSelectionTable")).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithQueryPatcher("courses.scope_by_role", CourseScopeByRole))

	lago.RegistryView.Register("courses.MultiSelectView",
		views.ListView[Course]("courses")(
			lago.GetPageView("courses.CourseMultiSelectionTable")).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithQueryPatcher("courses.scope_by_role", CourseScopeByRole).
			WithQueryPatcher("courses.multiselect_pool_course_ids", QueryPatcherMultiSelectPoolCourseIDs))
}
