package p_nirmancampus_assignmentsubmissions

import (
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/registry"
	"github.com/lariv-in/lago/views"
)

var assignmentSubmissionsAdminRoleMiddleware = p_users.RoleAuthorizationMiddleware{Roles: []string{"admin"}}

func init() {
	lago.RegistryView.Register("assignmentsubmissions.ListView",
		lago.GetPageView("assignmentsubmissions.Table").
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware{}).
			WithMiddleware("assignmentsubmissions.list", views.MiddlewareList[AssignmentSubmission]{
				Key: getters.Static("assignmentsubmissions"),
				QueryPatchers: views.QueryPatchers[AssignmentSubmission]{
					registry.Pair[string, views.QueryPatcher[AssignmentSubmission]]{Key: "assignmentsubmissions.preload_course", Value: views.QueryPatcherPreload[AssignmentSubmission]{Field: "Course"}},
					registry.Pair[string, views.QueryPatcher[AssignmentSubmission]]{Key: "assignmentsubmissions.scope_by_role", Value: AssignmentSubmissionScopeByRole},
				},
			}),
	)

	lago.RegistryView.Register("assignmentsubmissions.DetailView",
		lago.GetPageView("assignmentsubmissions.Detail").
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware{}).
			WithMiddleware("assignmentsubmissions.detail", views.MiddlewareDetail[AssignmentSubmission]{
				Key:          getters.Static("assignmentsubmission"),
				PathParamKey: getters.Static("id"),
				QueryPatchers: views.QueryPatchers[AssignmentSubmission]{
					registry.Pair[string, views.QueryPatcher[AssignmentSubmission]]{Key: "assignmentsubmissions.preload_course", Value: views.QueryPatcherPreload[AssignmentSubmission]{Field: "Course"}},
					registry.Pair[string, views.QueryPatcher[AssignmentSubmission]]{Key: "assignmentsubmissions.preload_academicrecord_student_user", Value: views.QueryPatcherPreload[AssignmentSubmission]{Field: "AcademicRecord.Student.User"}},
					registry.Pair[string, views.QueryPatcher[AssignmentSubmission]]{Key: "assignmentsubmissions.preload_academicrecord_program", Value: views.QueryPatcherPreload[AssignmentSubmission]{Field: "AcademicRecord.Program"}},
					registry.Pair[string, views.QueryPatcher[AssignmentSubmission]]{Key: "assignmentsubmissions.preload_assets", Value: views.QueryPatcherPreload[AssignmentSubmission]{Field: "Assets"}},
					registry.Pair[string, views.QueryPatcher[AssignmentSubmission]]{Key: "assignmentsubmissions.scope_by_role", Value: AssignmentSubmissionScopeByRole},
				},
			}),
	)

	lago.RegistryView.Register("assignmentsubmissions.CreateView",
		lago.GetPageView("assignmentsubmissions.CreateForm").
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware{}).
			WithMiddleware("assignmentsubmissions.admin_role", assignmentSubmissionsAdminRoleMiddleware).
			WithMiddleware("assignmentsubmissions.create_query_defaults", assignmentSubmissionCreateQueryDefaultsMiddleware{}).
			WithMiddleware("assignmentsubmissions.create", views.MiddlewareCreate[AssignmentSubmission]{
				SuccessURL: lago.RoutePath("assignmentsubmissions.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("$id")),
				}),
			}),
	)

	lago.RegistryView.Register("assignmentsubmissions.UpdateView",
		lago.GetPageView("assignmentsubmissions.UpdateForm").
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware{}).
			WithMiddleware("assignmentsubmissions.admin_role", assignmentSubmissionsAdminRoleMiddleware).
			WithMiddleware("assignmentsubmissions.detail", views.MiddlewareDetail[AssignmentSubmission]{
				Key:          getters.Static("assignmentsubmission"),
				PathParamKey: getters.Static("id"),
				QueryPatchers: views.QueryPatchers[AssignmentSubmission]{
					registry.Pair[string, views.QueryPatcher[AssignmentSubmission]]{Key: "assignmentsubmissions.preload_course", Value: views.QueryPatcherPreload[AssignmentSubmission]{Field: "Course"}},
					registry.Pair[string, views.QueryPatcher[AssignmentSubmission]]{Key: "assignmentsubmissions.preload_academicrecord_student_user", Value: views.QueryPatcherPreload[AssignmentSubmission]{Field: "AcademicRecord.Student.User"}},
					registry.Pair[string, views.QueryPatcher[AssignmentSubmission]]{Key: "assignmentsubmissions.preload_academicrecord_program", Value: views.QueryPatcherPreload[AssignmentSubmission]{Field: "AcademicRecord.Program"}},
					registry.Pair[string, views.QueryPatcher[AssignmentSubmission]]{Key: "assignmentsubmissions.preload_assets", Value: views.QueryPatcherPreload[AssignmentSubmission]{Field: "Assets"}},
					registry.Pair[string, views.QueryPatcher[AssignmentSubmission]]{Key: "assignmentsubmissions.scope_by_role", Value: AssignmentSubmissionScopeByRole},
				},
			}).
			WithMiddleware("assignmentsubmissions.update", views.MiddlewareUpdate[AssignmentSubmission]{
				Key: getters.Static("assignmentsubmission"),
				SuccessURL: lago.RoutePath("assignmentsubmissions.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("assignmentsubmission.ID")),
				}),
				QueryPatchers: views.QueryPatchers[AssignmentSubmission]{
					registry.Pair[string, views.QueryPatcher[AssignmentSubmission]]{Key: "assignmentsubmissions.scope_by_role", Value: AssignmentSubmissionScopeByRole},
				},
			}),
	)

	lago.RegistryView.Register("assignmentsubmissions.DeleteView",
		lago.GetPageView("assignmentsubmissions.DeleteForm").
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware{}).
			WithMiddleware("assignmentsubmissions.admin_role", assignmentSubmissionsAdminRoleMiddleware).
			WithMiddleware("assignmentsubmissions.detail", views.MiddlewareDetail[AssignmentSubmission]{
				Key:          getters.Static("assignmentsubmission"),
				PathParamKey: getters.Static("id"),
				QueryPatchers: views.QueryPatchers[AssignmentSubmission]{
					registry.Pair[string, views.QueryPatcher[AssignmentSubmission]]{Key: "assignmentsubmissions.preload_course", Value: views.QueryPatcherPreload[AssignmentSubmission]{Field: "Course"}},
					registry.Pair[string, views.QueryPatcher[AssignmentSubmission]]{Key: "assignmentsubmissions.scope_by_role", Value: AssignmentSubmissionScopeByRole},
				},
			}).
			WithMiddleware("assignmentsubmissions.delete", views.MiddlewareDelete[AssignmentSubmission]{
				Key:        getters.Static("assignmentsubmission"),
				SuccessURL: lago.RoutePath("assignmentsubmissions.DefaultRoute", nil),
				QueryPatchers: views.QueryPatchers[AssignmentSubmission]{
					registry.Pair[string, views.QueryPatcher[AssignmentSubmission]]{Key: "assignmentsubmissions.scope_by_role", Value: AssignmentSubmissionScopeByRole},
				},
			}),
	)
}
