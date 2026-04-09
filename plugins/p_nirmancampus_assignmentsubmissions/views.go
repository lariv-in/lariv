package p_nirmancampus_assignmentsubmissions

import (
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/registry"
	"github.com/lariv-in/lago/views"
)

var assignmentSubmissionsAdminRoleLayer = p_users.RoleAuthorizationLayer{Roles: []string{"admin"}}

func init() {
	lago.RegistryView.Register("assignmentsubmissions.ListView",
		lago.GetPageView("assignmentsubmissions.Table").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("assignmentsubmissions.list", views.LayerList[AssignmentSubmission]{
				Key: getters.Static("assignmentsubmissions"),
				QueryPatchers: views.QueryPatchers[AssignmentSubmission]{
					registry.Pair[string, views.QueryPatcher[AssignmentSubmission]]{Key: "assignmentsubmissions.preload_course", Value: views.QueryPatcherPreload[AssignmentSubmission]{Field: "Course"}},
					registry.Pair[string, views.QueryPatcher[AssignmentSubmission]]{Key: "assignmentsubmissions.scope_by_role", Value: AssignmentSubmissionScopeByRole},
				},
			}),
	)

	lago.RegistryView.Register("assignmentsubmissions.DetailView",
		lago.GetPageView("assignmentsubmissions.Detail").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("assignmentsubmissions.detail", views.LayerDetail[AssignmentSubmission]{
				Key:          getters.Static("assignmentsubmission"),
				PathParamKey: getters.Static("id"),
				QueryPatchers: views.QueryPatchers[AssignmentSubmission]{
					registry.Pair[string, views.QueryPatcher[AssignmentSubmission]]{Key: "assignmentsubmissions.preload_course", Value: views.QueryPatcherPreload[AssignmentSubmission]{Field: "Course"}},
					registry.Pair[string, views.QueryPatcher[AssignmentSubmission]]{Key: "assignmentsubmissions.preload_academicrecord_student", Value: views.QueryPatcherPreload[AssignmentSubmission]{Field: "AcademicRecord.Student"}},
					registry.Pair[string, views.QueryPatcher[AssignmentSubmission]]{Key: "assignmentsubmissions.preload_academicrecord_program", Value: views.QueryPatcherPreload[AssignmentSubmission]{Field: "AcademicRecord.Program"}},
					registry.Pair[string, views.QueryPatcher[AssignmentSubmission]]{Key: "assignmentsubmissions.preload_academicrecord_session", Value: views.QueryPatcherPreload[AssignmentSubmission]{Field: "AcademicRecord.Session"}},
					registry.Pair[string, views.QueryPatcher[AssignmentSubmission]]{Key: "assignmentsubmissions.preload_assets", Value: views.QueryPatcherPreload[AssignmentSubmission]{Field: "Assets"}},
					registry.Pair[string, views.QueryPatcher[AssignmentSubmission]]{Key: "assignmentsubmissions.scope_by_role", Value: AssignmentSubmissionScopeByRole},
				},
			}),
	)

	lago.RegistryView.Register("assignmentsubmissions.CreateView",
		lago.GetPageView("assignmentsubmissions.CreateForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("assignmentsubmissions.admin_role", assignmentSubmissionsAdminRoleLayer).
			WithLayer("assignmentsubmissions.create_query_defaults", assignmentSubmissionCreateQueryDefaultsLayer{}).
			WithLayer("assignmentsubmissions.create", views.LayerCreate[AssignmentSubmission]{
				SuccessURL: lago.RoutePath("assignmentsubmissions.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("$id")),
				}),
			}),
	)

	lago.RegistryView.Register("assignmentsubmissions.UpdateView",
		lago.GetPageView("assignmentsubmissions.UpdateForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("assignmentsubmissions.admin_role", assignmentSubmissionsAdminRoleLayer).
			WithLayer("assignmentsubmissions.detail", views.LayerDetail[AssignmentSubmission]{
				Key:          getters.Static("assignmentsubmission"),
				PathParamKey: getters.Static("id"),
				QueryPatchers: views.QueryPatchers[AssignmentSubmission]{
					registry.Pair[string, views.QueryPatcher[AssignmentSubmission]]{Key: "assignmentsubmissions.preload_course", Value: views.QueryPatcherPreload[AssignmentSubmission]{Field: "Course"}},
					registry.Pair[string, views.QueryPatcher[AssignmentSubmission]]{Key: "assignmentsubmissions.preload_academicrecord_student", Value: views.QueryPatcherPreload[AssignmentSubmission]{Field: "AcademicRecord.Student"}},
					registry.Pair[string, views.QueryPatcher[AssignmentSubmission]]{Key: "assignmentsubmissions.preload_academicrecord_program", Value: views.QueryPatcherPreload[AssignmentSubmission]{Field: "AcademicRecord.Program"}},
					registry.Pair[string, views.QueryPatcher[AssignmentSubmission]]{Key: "assignmentsubmissions.preload_academicrecord_session", Value: views.QueryPatcherPreload[AssignmentSubmission]{Field: "AcademicRecord.Session"}},
					registry.Pair[string, views.QueryPatcher[AssignmentSubmission]]{Key: "assignmentsubmissions.preload_assets", Value: views.QueryPatcherPreload[AssignmentSubmission]{Field: "Assets"}},
					registry.Pair[string, views.QueryPatcher[AssignmentSubmission]]{Key: "assignmentsubmissions.scope_by_role", Value: AssignmentSubmissionScopeByRole},
				},
			}).
			WithLayer("assignmentsubmissions.update", views.LayerUpdate[AssignmentSubmission]{
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
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("assignmentsubmissions.admin_role", assignmentSubmissionsAdminRoleLayer).
			WithLayer("assignmentsubmissions.detail", views.LayerDetail[AssignmentSubmission]{
				Key:          getters.Static("assignmentsubmission"),
				PathParamKey: getters.Static("id"),
				QueryPatchers: views.QueryPatchers[AssignmentSubmission]{
					registry.Pair[string, views.QueryPatcher[AssignmentSubmission]]{Key: "assignmentsubmissions.preload_course", Value: views.QueryPatcherPreload[AssignmentSubmission]{Field: "Course"}},
					registry.Pair[string, views.QueryPatcher[AssignmentSubmission]]{Key: "assignmentsubmissions.scope_by_role", Value: AssignmentSubmissionScopeByRole},
				},
			}).
			WithLayer("assignmentsubmissions.delete", views.LayerDelete[AssignmentSubmission]{
				Key:        getters.Static("assignmentsubmission"),
				SuccessURL: lago.RoutePath("assignmentsubmissions.DefaultRoute", nil),
				QueryPatchers: views.QueryPatchers[AssignmentSubmission]{
					registry.Pair[string, views.QueryPatcher[AssignmentSubmission]]{Key: "assignmentsubmissions.scope_by_role", Value: AssignmentSubmissionScopeByRole},
				},
			}),
	)
}
