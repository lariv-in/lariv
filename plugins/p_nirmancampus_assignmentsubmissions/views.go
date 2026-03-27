package p_nirmancampus_assignmentsubmissions

import (
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/views"
)

var assignmentSubmissionsAdminRoleMiddleware = p_users.RoleAuthorizationMiddleware([]string{"admin"})

func init() {
	lago.RegistryView.Register("assignmentsubmissions.ListView",
		views.ListView[AssignmentSubmission]("assignmentsubmissions")(
			lago.GetPageView("assignmentsubmissions.Table"),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithQueryPatcher("assignmentsubmissions.preload_course", views.QueryPatcherPreload("Course")).
			WithQueryPatcher("assignmentsubmissions.scope_by_role", AssignmentSubmissionScopeByRole),
	)

	lago.RegistryView.Register("assignmentsubmissions.DetailView",
		views.DetailView[AssignmentSubmission]("assignmentsubmission")(
			lago.GetPageView("assignmentsubmissions.Detail"),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithQueryPatcher("assignmentsubmissions.preload_course", views.QueryPatcherPreload("Course")).
			WithQueryPatcher("assignmentsubmissions.preload_academicrecord_student_user", views.QueryPatcherPreload("AcademicRecord.Student.User")).
			WithQueryPatcher("assignmentsubmissions.preload_academicrecord_program", views.QueryPatcherPreload("AcademicRecord.Program")).
			WithQueryPatcher("assignmentsubmissions.preload_assets", views.QueryPatcherPreload("Assets")).
			WithQueryPatcher("assignmentsubmissions.scope_by_role", AssignmentSubmissionScopeByRole),
	)

	lago.RegistryView.Register("assignmentsubmissions.CreateView",
		views.CreateView[AssignmentSubmission](
			lago.GetterRoutePath("assignmentsubmissions.DetailRoute", map[string]getters.Getter[any]{
				"id": getters.GetterAny(getters.GetterKey[uint]("$id")),
			}),
		)(
			lago.GetPageView("assignmentsubmissions.CreateForm"),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithMiddleware("assignmentsubmissions.admin_role", assignmentSubmissionsAdminRoleMiddleware),
	)

	lago.RegistryView.Register("assignmentsubmissions.UpdateView",
		views.DetailView[AssignmentSubmission]("assignmentsubmission")(
			views.UpdateView[AssignmentSubmission](
				lago.GetterRoutePath("assignmentsubmissions.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("$id")),
				}),
			)(
				lago.GetPageView("assignmentsubmissions.UpdateForm"),
			),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithMiddleware("assignmentsubmissions.admin_role", assignmentSubmissionsAdminRoleMiddleware).
			WithQueryPatcher("assignmentsubmissions.preload_course", views.QueryPatcherPreload("Course")).
			WithQueryPatcher("assignmentsubmissions.preload_academicrecord_student_user", views.QueryPatcherPreload("AcademicRecord.Student.User")).
			WithQueryPatcher("assignmentsubmissions.preload_academicrecord_program", views.QueryPatcherPreload("AcademicRecord.Program")).
			WithQueryPatcher("assignmentsubmissions.preload_assets", views.QueryPatcherPreload("Assets")).
			WithQueryPatcher("assignmentsubmissions.scope_by_role", AssignmentSubmissionScopeByRole),
	)

	lago.RegistryView.Register("assignmentsubmissions.DeleteView",
		views.DetailView[AssignmentSubmission]("assignmentsubmission")(
			views.DeleteView[AssignmentSubmission](
				lago.GetterRoutePath("assignmentsubmissions.DefaultRoute", nil),
			)(
				lago.GetPageView("assignmentsubmissions.DeleteForm"),
			),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithMiddleware("assignmentsubmissions.admin_role", assignmentSubmissionsAdminRoleMiddleware).
			WithQueryPatcher("assignmentsubmissions.preload_course", views.QueryPatcherPreload("Course")).
			WithQueryPatcher("assignmentsubmissions.scope_by_role", AssignmentSubmissionScopeByRole),
	)
}
