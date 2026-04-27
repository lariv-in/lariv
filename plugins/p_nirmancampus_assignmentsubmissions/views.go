package p_nirmancampus_assignmentsubmissions

import (
	"net/http"

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
					registry.Pair[string, views.QueryPatcher[AssignmentSubmission]]{
						Key:   "assignmentsubmissions.preload",
						Value: views.QueryPatcherPreload[AssignmentSubmission]{Fields: []string{"Course", "AcademicRecord.Student", "AcademicRecord.AdmissionSession"}},
					},
					registry.Pair[string, views.QueryPatcher[AssignmentSubmission]]{
						Key:   "assignmentsubmissions.filter_by_session",
						Value: AssignmentSubmissionListSessionFilter,
					},
					registry.Pair[string, views.QueryPatcher[AssignmentSubmission]]{
						Key:   "assignmentsubmissions.scope_by_role",
						Value: AssignmentSubmissionScopeByRole,
					},
				},
			}).
			WithLayer("assignmentsubmissions.list_filter_academic_record", listFilterAcademicRecordLoadLayer{}),
	)

	lago.RegistryView.Register("assignmentsubmissions.DetailView",
		lago.GetPageView("assignmentsubmissions.Detail").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("assignmentsubmissions.detail", views.LayerDetail[AssignmentSubmission]{
				Key:          getters.Static("assignmentsubmission"),
				PathParamKey: getters.Static("id"),
				QueryPatchers: views.QueryPatchers[AssignmentSubmission]{
					registry.Pair[string, views.QueryPatcher[AssignmentSubmission]]{Key: "assignmentsubmissions.preload", Value: views.QueryPatcherPreload[AssignmentSubmission]{Fields: []string{"Course", "AcademicRecord.Student", "AcademicRecord.Program", "AcademicRecord.AdmissionSession", "Assets"}}},
					registry.Pair[string, views.QueryPatcher[AssignmentSubmission]]{Key: "assignmentsubmissions.scope_by_role", Value: AssignmentSubmissionScopeByRole},
				},
			}),
	)

	lago.RegistryView.Register("assignmentsubmissions.BulkCreateFromAcademicRecordView",
		lago.GetPageView("assignmentsubmissions.BulkCreateFromAcademicRecordForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("assignmentsubmissions.admin_role", assignmentSubmissionsAdminRoleLayer).
			WithLayer("assignmentsubmissions.bulk_academic_record_load", academicRecordBulkCreateLoadLayer{}).
			WithLayer("assignmentsubmissions.bulk_academic_record_post", views.MethodLayer{
				Method:  http.MethodPost,
				Handler: bulkCreateFromAcademicRecordPostHandler,
			}),
	)

	lago.RegistryView.Register("assignmentsubmissions.BulkAddMarksFromAcademicRecordView",
		lago.GetPageView("assignmentsubmissions.BulkAddMarksFromAcademicRecordForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("assignmentsubmissions.admin_role", assignmentSubmissionsAdminRoleLayer).
			WithLayer("assignmentsubmissions.bulk_add_marks_load", academicRecordBulkAddMarksLoadLayer{}).
			WithLayer("assignmentsubmissions.bulk_add_marks_post", views.MethodLayer{
				Method:  http.MethodPost,
				Handler: bulkAddMarksFromAcademicRecordPostHandler,
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
					registry.Pair[string, views.QueryPatcher[AssignmentSubmission]]{Key: "assignmentsubmissions.preload", Value: views.QueryPatcherPreload[AssignmentSubmission]{Fields: []string{"Course", "AcademicRecord.Student", "AcademicRecord.Program", "AcademicRecord.AdmissionSession", "Assets"}}},
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
					registry.Pair[string, views.QueryPatcher[AssignmentSubmission]]{Key: "assignmentsubmissions.preload", Value: views.QueryPatcherPreload[AssignmentSubmission]{Fields: []string{"Course"}}},
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
