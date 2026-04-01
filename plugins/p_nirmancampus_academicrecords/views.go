package p_nirmancampus_academicrecords

import (
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/views"
)

// academicRecordsAdminRoleMiddleware restricts create/update/delete to admin;
// superusers are always allowed (see p_users.RoleAuthorizationMiddleware).
var academicRecordsAdminRoleMiddleware = p_users.RoleAuthorizationMiddleware([]string{"admin"})

func init() {
	// List view
	lago.RegistryView.Register("academicrecords.ListView",
		views.ListView[AcademicRecord]("academicrecords")(
			lago.GetPageView("academicrecords.AcademicRecordTable"),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithQueryPatcher("academicrecords.preload_student_user", views.QueryPatcherPreload("Student.User")).
			WithQueryPatcher("academicrecords.preload_program", views.QueryPatcherPreload("Program")).
			WithQueryPatcher("academicrecords.scope_by_role", AcademicRecordScopeByRole),
	)

	// Detail view
	lago.RegistryView.Register("academicrecords.DetailView",
		views.DetailView[AcademicRecord]("academicrecord", "id")(
			lago.GetPageView("academicrecords.AcademicRecordDetail"),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithQueryPatcher("academicrecords.preload_student_user", views.QueryPatcherPreload("Student.User")).
			WithQueryPatcher("academicrecords.preload_program", views.QueryPatcherPreload("Program")).
			WithQueryPatcher("academicrecords.preload_compulsory_courses", views.QueryPatcherPreload("CompulsoryCourses")).
			WithQueryPatcher("academicrecords.preload_optional_courses", views.QueryPatcherPreload("OptionalCourses")).
			WithQueryPatcher("academicrecords.scope_by_role", AcademicRecordScopeByRole),
	)

	// Create view
	lago.RegistryView.Register("academicrecords.CreateView",
		views.CreateView[AcademicRecord](
			lago.GetterRoutePath("academicrecords.DetailRoute", map[string]getters.Getter[any]{
				"id": getters.Any(getters.Key[uint]("$id")),
			}),
		)(
			lago.GetPageView("academicrecords.AcademicRecordCreateForm"),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithMiddleware("academicrecords.admin_role", academicRecordsAdminRoleMiddleware).
			WithFormPatcher("academicrecords.create_from_program_structure", formPatcherAcademicRecordCreateFromProgramStructure),
	)

	// Update view
	lago.RegistryView.Register("academicrecords.UpdateView",
		views.DetailView[AcademicRecord]("academicrecord", "id")(
			views.UpdateView[AcademicRecord]("id",
				lago.GetterRoutePath("academicrecords.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("$id")),
				}),
			)(
				lago.GetPageView("academicrecords.AcademicRecordUpdateForm"),
			),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithMiddleware("academicrecords.admin_role", academicRecordsAdminRoleMiddleware).
			WithQueryPatcher("academicrecords.preload_student_user", views.QueryPatcherPreload("Student.User")).
			WithQueryPatcher("academicrecords.preload_program", views.QueryPatcherPreload("Program")).
			WithQueryPatcher("academicrecords.preload_compulsory_courses", views.QueryPatcherPreload("CompulsoryCourses")).
			WithQueryPatcher("academicrecords.preload_optional_courses", views.QueryPatcherPreload("OptionalCourses")).
			WithQueryPatcher("academicrecords.scope_by_role", AcademicRecordScopeByRole).
			WithFormValidator("academicrecords.optional_course_count", formValidatorAcademicRecordOptionalCourseCount),
	)

	// Delete view
	lago.RegistryView.Register("academicrecords.DeleteView",
		views.DetailView[AcademicRecord]("academicrecord", "id")(
			views.DeleteView[AcademicRecord]("id",
				lago.GetterRoutePath("academicrecords.DefaultRoute", nil),
			)(
				lago.GetPageView("academicrecords.AcademicRecordDeleteForm"),
			),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithMiddleware("academicrecords.admin_role", academicRecordsAdminRoleMiddleware).
			WithQueryPatcher("academicrecords.preload_student_user", views.QueryPatcherPreload("Student.User")).
			WithQueryPatcher("academicrecords.preload_program", views.QueryPatcherPreload("Program")).
			WithQueryPatcher("academicrecords.preload_compulsory_courses", views.QueryPatcherPreload("CompulsoryCourses")).
			WithQueryPatcher("academicrecords.preload_optional_courses", views.QueryPatcherPreload("OptionalCourses")).
			WithQueryPatcher("academicrecords.scope_by_role", AcademicRecordScopeByRole),
	)

	// Selection view
	lago.RegistryView.Register("academicrecords.SelectView",
		views.ListView[AcademicRecord]("academicrecords")(
			lago.GetPageView("academicrecords.AcademicRecordSelectionTable"),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithQueryPatcher("academicrecords.preload_student_user", views.QueryPatcherPreload("Student.User")).
			WithQueryPatcher("academicrecords.preload_program", views.QueryPatcherPreload("Program")).
			WithQueryPatcher("academicrecords.scope_by_role", AcademicRecordScopeByRole),
	)
}
