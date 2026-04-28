package p_nirmancampus_examregistrations

import (
	"net/http"

	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/registry"
	"github.com/lariv-in/lago/views"
)

var examRegistrationsAdminRoleLayer = p_users.RoleAuthorizationLayer{Roles: []string{"admin"}}

func init() {
	lago.RegistryView.Register("examregistrations.ListView",
		lago.GetPageView("examregistrations.Table").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("examregistrations.list", views.LayerList[ExamRegistration]{
				Key: getters.Static("examregistrations"),
				QueryPatchers: views.QueryPatchers[ExamRegistration]{
					registry.Pair[string, views.QueryPatcher[ExamRegistration]]{
						Key:   "examregistrations.preload",
						Value: views.QueryPatcherPreload[ExamRegistration]{Fields: []string{"Course", "AcademicRecord.Student", "AcademicRecord.AdmissionSession"}},
					},
					registry.Pair[string, views.QueryPatcher[ExamRegistration]]{
						Key:   "examregistrations.filter_by_session",
						Value: ExamRegistrationListSessionFilter,
					},
					registry.Pair[string, views.QueryPatcher[ExamRegistration]]{
						Key:   "examregistrations.scope_by_role",
						Value: ExamRegistrationScopeByRole,
					},
					registry.Pair[string, views.QueryPatcher[ExamRegistration]]{
						Key:   "examregistrations.list_order",
						Value: ExamRegistrationListOrder,
					},
				},
			}).
			WithLayer("examregistrations.list_filter_academic_record", listFilterAcademicRecordLoadLayer{}),
	)

	lago.RegistryView.Register("examregistrations.DetailView",
		lago.GetPageView("examregistrations.Detail").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("examregistrations.detail", views.LayerDetail[ExamRegistration]{
				Key:          getters.Static("examregistration"),
				PathParamKey: getters.Static("id"),
				QueryPatchers: views.QueryPatchers[ExamRegistration]{
					registry.Pair[string, views.QueryPatcher[ExamRegistration]]{Key: "examregistrations.preload", Value: views.QueryPatcherPreload[ExamRegistration]{Fields: []string{"Course", "AcademicRecord.Student", "AcademicRecord.Program", "AcademicRecord.AdmissionSession", "Assets"}}},
					registry.Pair[string, views.QueryPatcher[ExamRegistration]]{Key: "examregistrations.scope_by_role", Value: ExamRegistrationScopeByRole},
				},
			}),
	)

	lago.RegistryView.Register("examregistrations.BulkCreateFromAcademicRecordView",
		lago.GetPageView("examregistrations.BulkCreateFromAcademicRecordForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("examregistrations.admin_role", examRegistrationsAdminRoleLayer).
			WithLayer("examregistrations.bulk_academic_record_load", academicRecordBulkCreateLoadLayer{}).
			WithLayer("examregistrations.bulk_academic_record_post", views.MethodLayer{
				Method:  http.MethodPost,
				Handler: bulkCreateFromAcademicRecordPostHandler,
			}),
	)

	lago.RegistryView.Register("examregistrations.UpdateView",
		lago.GetPageView("examregistrations.UpdateForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("examregistrations.admin_role", examRegistrationsAdminRoleLayer).
			WithLayer("examregistrations.detail", views.LayerDetail[ExamRegistration]{
				Key:          getters.Static("examregistration"),
				PathParamKey: getters.Static("id"),
				QueryPatchers: views.QueryPatchers[ExamRegistration]{
					registry.Pair[string, views.QueryPatcher[ExamRegistration]]{Key: "examregistrations.preload", Value: views.QueryPatcherPreload[ExamRegistration]{Fields: []string{"Course", "AcademicRecord.Student", "AcademicRecord.Program", "AcademicRecord.AdmissionSession", "Assets"}}},
					registry.Pair[string, views.QueryPatcher[ExamRegistration]]{Key: "examregistrations.scope_by_role", Value: ExamRegistrationScopeByRole},
				},
			}).
			WithLayer("examregistrations.update", views.LayerUpdate[ExamRegistration]{
				Key: getters.Static("examregistration"),
				SuccessURL: lago.RoutePath("examregistrations.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("examregistration.ID")),
				}),
				QueryPatchers: views.QueryPatchers[ExamRegistration]{
					registry.Pair[string, views.QueryPatcher[ExamRegistration]]{Key: "examregistrations.scope_by_role", Value: ExamRegistrationScopeByRole},
				},
			}),
	)

	lago.RegistryView.Register("examregistrations.DeleteView",
		lago.GetPageView("examregistrations.DeleteForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("examregistrations.admin_role", examRegistrationsAdminRoleLayer).
			WithLayer("examregistrations.detail", views.LayerDetail[ExamRegistration]{
				Key:          getters.Static("examregistration"),
				PathParamKey: getters.Static("id"),
				QueryPatchers: views.QueryPatchers[ExamRegistration]{
					registry.Pair[string, views.QueryPatcher[ExamRegistration]]{Key: "examregistrations.preload", Value: views.QueryPatcherPreload[ExamRegistration]{Fields: []string{"Course"}}},
					registry.Pair[string, views.QueryPatcher[ExamRegistration]]{Key: "examregistrations.scope_by_role", Value: ExamRegistrationScopeByRole},
				},
			}).
			WithLayer("examregistrations.delete", views.LayerDelete[ExamRegistration]{
				Key:        getters.Static("examregistration"),
				SuccessURL: lago.RoutePath("examregistrations.DefaultRoute", nil),
				QueryPatchers: views.QueryPatchers[ExamRegistration]{
					registry.Pair[string, views.QueryPatcher[ExamRegistration]]{Key: "examregistrations.scope_by_role", Value: ExamRegistrationScopeByRole},
				},
			}),
	)

	lago.RegistryView.Register("examregistrations.AcademicRecordExamReceiptView",
		lago.GetPageView("academicrecords.AcademicRecordDetail").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("examregistrations.academic_record_exam_receipt", views.MethodLayer{
				Method:  http.MethodGet,
				Handler: downloadAcademicRecordExamReceiptHandler,
			}),
	)
}
