package p_nirmancampus_academicrecords

import (
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/registry"
	"github.com/lariv-in/lago/views"
)

// academicRecordsAdminRoleMiddleware restricts create/update/delete to admin;
// superusers are always allowed (see p_users.RoleAuthorizationMiddleware).
var academicRecordsAdminRoleMiddleware = p_users.RoleAuthorizationMiddleware{Roles: []string{"admin"}}

func init() {
	// List view
	lago.RegistryView.Register("academicrecords.ListView",
		lago.GetPageView("academicrecords.AcademicRecordTable").
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware{}).
			WithMiddleware("academicrecords.list", views.MiddlewareList[AcademicRecord]{
				Key: getters.Static("academicrecords"),
				QueryPatchers: views.QueryPatchers[AcademicRecord]{
					registry.Pair[string, views.QueryPatcher[AcademicRecord]]{Key: "academicrecords.preload_student_user", Value: views.QueryPatcherPreload[AcademicRecord]{Field: "Student.User"}},
					registry.Pair[string, views.QueryPatcher[AcademicRecord]]{Key: "academicrecords.preload_program", Value: views.QueryPatcherPreload[AcademicRecord]{Field: "Program"}},
					registry.Pair[string, views.QueryPatcher[AcademicRecord]]{Key: "academicrecords.scope_by_role", Value: AcademicRecordScopeByRole},
				},
			}),
	)

	// Detail view
	lago.RegistryView.Register("academicrecords.DetailView",
		lago.GetPageView("academicrecords.AcademicRecordDetail").
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware{}).
			WithMiddleware("academicrecords.detail", views.MiddlewareDetail[AcademicRecord]{
				Key:          getters.Static("academicrecord"),
				PathParamKey: getters.Static("id"),
				QueryPatchers: views.QueryPatchers[AcademicRecord]{
					registry.Pair[string, views.QueryPatcher[AcademicRecord]]{Key: "academicrecords.preload_student_user", Value: views.QueryPatcherPreload[AcademicRecord]{Field: "Student.User"}},
					registry.Pair[string, views.QueryPatcher[AcademicRecord]]{Key: "academicrecords.preload_program", Value: views.QueryPatcherPreload[AcademicRecord]{Field: "Program"}},
					registry.Pair[string, views.QueryPatcher[AcademicRecord]]{Key: "academicrecords.preload_compulsory_courses", Value: views.QueryPatcherPreload[AcademicRecord]{Field: "CompulsoryCourses"}},
					registry.Pair[string, views.QueryPatcher[AcademicRecord]]{Key: "academicrecords.preload_optional_courses", Value: views.QueryPatcherPreload[AcademicRecord]{Field: "OptionalCourses"}},
					registry.Pair[string, views.QueryPatcher[AcademicRecord]]{Key: "academicrecords.scope_by_role", Value: AcademicRecordScopeByRole},
				},
			}),
	)

	// Create view
	lago.RegistryView.Register("academicrecords.CreateView",
		lago.GetPageView("academicrecords.AcademicRecordCreateForm").
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware{}).
			WithMiddleware("academicrecords.admin_role", academicRecordsAdminRoleMiddleware).
			WithMiddleware("academicrecords.create", views.MiddlewareCreate[AcademicRecord]{
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
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware{}).
			WithMiddleware("academicrecords.admin_role", academicRecordsAdminRoleMiddleware).
			WithMiddleware("academicrecords.detail", views.MiddlewareDetail[AcademicRecord]{
				Key:          getters.Static("academicrecord"),
				PathParamKey: getters.Static("id"),
				QueryPatchers: views.QueryPatchers[AcademicRecord]{
					registry.Pair[string, views.QueryPatcher[AcademicRecord]]{Key: "academicrecords.preload_student_user", Value: views.QueryPatcherPreload[AcademicRecord]{Field: "Student.User"}},
					registry.Pair[string, views.QueryPatcher[AcademicRecord]]{Key: "academicrecords.preload_program", Value: views.QueryPatcherPreload[AcademicRecord]{Field: "Program"}},
					registry.Pair[string, views.QueryPatcher[AcademicRecord]]{Key: "academicrecords.preload_compulsory_courses", Value: views.QueryPatcherPreload[AcademicRecord]{Field: "CompulsoryCourses"}},
					registry.Pair[string, views.QueryPatcher[AcademicRecord]]{Key: "academicrecords.preload_optional_courses", Value: views.QueryPatcherPreload[AcademicRecord]{Field: "OptionalCourses"}},
					registry.Pair[string, views.QueryPatcher[AcademicRecord]]{Key: "academicrecords.scope_by_role", Value: AcademicRecordScopeByRole},
				},
			}).
			WithMiddleware("academicrecords.program_structure_unit", academicRecordProgramStructureUnitContextMiddleware{}).
			WithMiddleware("academicrecords.update", views.MiddlewareUpdate[AcademicRecord]{
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
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware{}).
			WithMiddleware("academicrecords.admin_role", academicRecordsAdminRoleMiddleware).
			WithMiddleware("academicrecords.detail", views.MiddlewareDetail[AcademicRecord]{
				Key:          getters.Static("academicrecord"),
				PathParamKey: getters.Static("id"),
				QueryPatchers: views.QueryPatchers[AcademicRecord]{
					registry.Pair[string, views.QueryPatcher[AcademicRecord]]{Key: "academicrecords.preload_student_user", Value: views.QueryPatcherPreload[AcademicRecord]{Field: "Student.User"}},
					registry.Pair[string, views.QueryPatcher[AcademicRecord]]{Key: "academicrecords.preload_program", Value: views.QueryPatcherPreload[AcademicRecord]{Field: "Program"}},
					registry.Pair[string, views.QueryPatcher[AcademicRecord]]{Key: "academicrecords.preload_compulsory_courses", Value: views.QueryPatcherPreload[AcademicRecord]{Field: "CompulsoryCourses"}},
					registry.Pair[string, views.QueryPatcher[AcademicRecord]]{Key: "academicrecords.preload_optional_courses", Value: views.QueryPatcherPreload[AcademicRecord]{Field: "OptionalCourses"}},
					registry.Pair[string, views.QueryPatcher[AcademicRecord]]{Key: "academicrecords.scope_by_role", Value: AcademicRecordScopeByRole},
				},
			}).
			WithMiddleware("academicrecords.delete", views.MiddlewareDelete[AcademicRecord]{
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
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware{}).
			WithMiddleware("academicrecords.select", views.MiddlewareList[AcademicRecord]{
				Key: getters.Static("academicrecords"),
				QueryPatchers: views.QueryPatchers[AcademicRecord]{
					registry.Pair[string, views.QueryPatcher[AcademicRecord]]{Key: "academicrecords.preload_student_user", Value: views.QueryPatcherPreload[AcademicRecord]{Field: "Student.User"}},
					registry.Pair[string, views.QueryPatcher[AcademicRecord]]{Key: "academicrecords.preload_program", Value: views.QueryPatcherPreload[AcademicRecord]{Field: "Program"}},
					registry.Pair[string, views.QueryPatcher[AcademicRecord]]{Key: "academicrecords.scope_by_role", Value: AcademicRecordScopeByRole},
				},
			}),
	)
}
