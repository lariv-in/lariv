package p_nirmancampus_academicrecords

import (
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/registry"
	"github.com/lariv-in/lago/views"
)

// academicRecordsAdminRoleLayer restricts create/update/delete to admin;
// superusers are always allowed (see p_users.RoleAuthorizationLayer).
var academicRecordsAdminRoleLayer = p_users.RoleAuthorizationLayer{Roles: []string{"admin"}}

func init() {
	// List view
	lago.RegistryView.Register("academicrecords.ListView",
		lago.GetPageView("academicrecords.AcademicRecordTable").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("academicrecords.list", views.LayerList[AcademicRecord]{
				Key: getters.Static("academicrecords"),
				QueryPatchers: views.QueryPatchers[AcademicRecord]{
					registry.Pair[string, views.QueryPatcher[AcademicRecord]]{Key: "academicrecords.preload_student", Value: views.QueryPatcherPreload[AcademicRecord]{Field: "Student"}},
					registry.Pair[string, views.QueryPatcher[AcademicRecord]]{Key: "academicrecords.preload_program", Value: views.QueryPatcherPreload[AcademicRecord]{Field: "Program"}},
					registry.Pair[string, views.QueryPatcher[AcademicRecord]]{Key: "academicrecords.preload_session", Value: views.QueryPatcherPreload[AcademicRecord]{Field: "Session"}},
					registry.Pair[string, views.QueryPatcher[AcademicRecord]]{Key: "academicrecords.filter_by_session", Value: AcademicRecordListSessionFilter},
					registry.Pair[string, views.QueryPatcher[AcademicRecord]]{Key: "academicrecords.scope_by_role", Value: AcademicRecordScopeByRole},
				},
			}),
	)

	// Detail view
	lago.RegistryView.Register("academicrecords.DetailView",
		lago.GetPageView("academicrecords.AcademicRecordDetail").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("academicrecords.detail", views.LayerDetail[AcademicRecord]{
				Key:          getters.Static("academicrecord"),
				PathParamKey: getters.Static("id"),
				QueryPatchers: views.QueryPatchers[AcademicRecord]{
					registry.Pair[string, views.QueryPatcher[AcademicRecord]]{Key: "academicrecords.preload_student", Value: views.QueryPatcherPreload[AcademicRecord]{Field: "Student"}},
					registry.Pair[string, views.QueryPatcher[AcademicRecord]]{Key: "academicrecords.preload_program", Value: views.QueryPatcherPreload[AcademicRecord]{Field: "Program"}},
					registry.Pair[string, views.QueryPatcher[AcademicRecord]]{Key: "academicrecords.preload_session", Value: views.QueryPatcherPreload[AcademicRecord]{Field: "Session"}},
					registry.Pair[string, views.QueryPatcher[AcademicRecord]]{Key: "academicrecords.preload_compulsory_courses", Value: views.QueryPatcherPreload[AcademicRecord]{Field: "CompulsoryCourses"}},
					registry.Pair[string, views.QueryPatcher[AcademicRecord]]{Key: "academicrecords.preload_optional_courses", Value: views.QueryPatcherPreload[AcademicRecord]{Field: "OptionalCourses"}},
					registry.Pair[string, views.QueryPatcher[AcademicRecord]]{Key: "academicrecords.scope_by_role", Value: AcademicRecordScopeByRole},
				},
			}),
	)

	// Create view
	lago.RegistryView.Register("academicrecords.CreateView",
		lago.GetPageView("academicrecords.AcademicRecordCreateForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("academicrecords.admin_role", academicRecordsAdminRoleLayer).
			WithLayer("academicrecords.create_query_defaults", academicRecordCreateQueryDefaultsLayer{}).
			WithLayer("academicrecords.create", views.LayerCreate[AcademicRecord]{
				SuccessURL: lago.RoutePath("academicrecords.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("$id")),
				}),
				FormPatchers: views.FormPatchers{
					registry.Pair[string, views.FormPatcher]{Key: "academicrecords.create_from_program_structure", Value: formPatcherAcademicRecordCreate{}},
				},
			}),
	)

	// Update view
	lago.RegistryView.Register("academicrecords.UpdateView",
		lago.GetPageView("academicrecords.AcademicRecordUpdateForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("academicrecords.admin_role", academicRecordsAdminRoleLayer).
			WithLayer("academicrecords.detail", views.LayerDetail[AcademicRecord]{
				Key:          getters.Static("academicrecord"),
				PathParamKey: getters.Static("id"),
				QueryPatchers: views.QueryPatchers[AcademicRecord]{
					registry.Pair[string, views.QueryPatcher[AcademicRecord]]{Key: "academicrecords.preload_student", Value: views.QueryPatcherPreload[AcademicRecord]{Field: "Student"}},
					registry.Pair[string, views.QueryPatcher[AcademicRecord]]{Key: "academicrecords.preload_program", Value: views.QueryPatcherPreload[AcademicRecord]{Field: "Program"}},
					registry.Pair[string, views.QueryPatcher[AcademicRecord]]{Key: "academicrecords.preload_session", Value: views.QueryPatcherPreload[AcademicRecord]{Field: "Session"}},
					registry.Pair[string, views.QueryPatcher[AcademicRecord]]{Key: "academicrecords.preload_compulsory_courses", Value: views.QueryPatcherPreload[AcademicRecord]{Field: "CompulsoryCourses"}},
					registry.Pair[string, views.QueryPatcher[AcademicRecord]]{Key: "academicrecords.preload_optional_courses", Value: views.QueryPatcherPreload[AcademicRecord]{Field: "OptionalCourses"}},
					registry.Pair[string, views.QueryPatcher[AcademicRecord]]{Key: "academicrecords.scope_by_role", Value: AcademicRecordScopeByRole},
				},
			}).
			WithLayer("academicrecords.program_structure_unit", academicRecordProgramStructureUnitContextLayer{}).
			WithLayer("academicrecords.update", views.LayerUpdate[AcademicRecord]{
				Key: getters.Static("academicrecord"),
				SuccessURL: lago.RoutePath("academicrecords.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("academicrecord.ID")),
				}),
				QueryPatchers: views.QueryPatchers[AcademicRecord]{
					registry.Pair[string, views.QueryPatcher[AcademicRecord]]{Key: "academicrecords.scope_by_role", Value: AcademicRecordScopeByRole},
				},
				FormPatchers: views.FormPatchers{
					registry.Pair[string, views.FormPatcher]{Key: "academicrecords.optional_course_count", Value: formPatcherAcademicRecordUpdate{}},
				},
			}),
	)

	// Delete view
	lago.RegistryView.Register("academicrecords.DeleteView",
		lago.GetPageView("academicrecords.AcademicRecordDeleteForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("academicrecords.admin_role", academicRecordsAdminRoleLayer).
			WithLayer("academicrecords.detail", views.LayerDetail[AcademicRecord]{
				Key:          getters.Static("academicrecord"),
				PathParamKey: getters.Static("id"),
				QueryPatchers: views.QueryPatchers[AcademicRecord]{
					registry.Pair[string, views.QueryPatcher[AcademicRecord]]{Key: "academicrecords.preload_student", Value: views.QueryPatcherPreload[AcademicRecord]{Field: "Student"}},
					registry.Pair[string, views.QueryPatcher[AcademicRecord]]{Key: "academicrecords.preload_program", Value: views.QueryPatcherPreload[AcademicRecord]{Field: "Program"}},
					registry.Pair[string, views.QueryPatcher[AcademicRecord]]{Key: "academicrecords.preload_session", Value: views.QueryPatcherPreload[AcademicRecord]{Field: "Session"}},
					registry.Pair[string, views.QueryPatcher[AcademicRecord]]{Key: "academicrecords.preload_compulsory_courses", Value: views.QueryPatcherPreload[AcademicRecord]{Field: "CompulsoryCourses"}},
					registry.Pair[string, views.QueryPatcher[AcademicRecord]]{Key: "academicrecords.preload_optional_courses", Value: views.QueryPatcherPreload[AcademicRecord]{Field: "OptionalCourses"}},
					registry.Pair[string, views.QueryPatcher[AcademicRecord]]{Key: "academicrecords.scope_by_role", Value: AcademicRecordScopeByRole},
				},
			}).
			WithLayer("academicrecords.delete", views.LayerDelete[AcademicRecord]{
				Key:        getters.Static("academicrecord"),
				SuccessURL: lago.RoutePath("academicrecords.DefaultRoute", nil),
				QueryPatchers: views.QueryPatchers[AcademicRecord]{
					registry.Pair[string, views.QueryPatcher[AcademicRecord]]{Key: "academicrecords.scope_by_role", Value: AcademicRecordScopeByRole},
				},
			}),
	)

	// Selection view
	lago.RegistryView.Register("academicrecords.SelectView",
		lago.GetPageView("academicrecords.AcademicRecordSelectionTable").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("academicrecords.select", views.LayerList[AcademicRecord]{
				Key: getters.Static("academicrecords"),
				QueryPatchers: views.QueryPatchers[AcademicRecord]{
					registry.Pair[string, views.QueryPatcher[AcademicRecord]]{Key: "academicrecords.preload_student", Value: views.QueryPatcherPreload[AcademicRecord]{Field: "Student"}},
					registry.Pair[string, views.QueryPatcher[AcademicRecord]]{Key: "academicrecords.preload_program", Value: views.QueryPatcherPreload[AcademicRecord]{Field: "Program"}},
					registry.Pair[string, views.QueryPatcher[AcademicRecord]]{Key: "academicrecords.preload_session", Value: views.QueryPatcherPreload[AcademicRecord]{Field: "Session"}},
					registry.Pair[string, views.QueryPatcher[AcademicRecord]]{Key: "academicrecords.filter_by_session", Value: AcademicRecordListSessionFilter},
					registry.Pair[string, views.QueryPatcher[AcademicRecord]]{Key: "academicrecords.scope_by_role", Value: AcademicRecordScopeByRole},
				},
			}),
	)
}
