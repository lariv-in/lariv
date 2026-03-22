package p_academicrecords

import (
	"github.com/lariv-in/getters"
	"github.com/lariv-in/lago"
	"github.com/lariv-in/p_users"
	"github.com/lariv-in/views"
)

func init() {
	// List view
	lago.RegistryView.Register("academicrecords.ListView",
		views.ListView[AcademicRecord]("academicrecords")(
			lago.GetPageView("academicrecords.AcademicRecordTable"),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithQueryPatcher("academicrecords.preload_student_user", views.QueryPatcherPreload("Student.User")).
			WithQueryPatcher("academicrecords.preload_semester", views.QueryPatcherPreload("Semester")).
			WithQueryPatcher("academicrecords.scope_by_role", AcademicRecordScopeByRole),
	)

	// Detail view
	lago.RegistryView.Register("academicrecords.DetailView",
		views.DetailView[AcademicRecord]("academicrecord")(
			lago.GetPageView("academicrecords.AcademicRecordDetail"),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithQueryPatcher("academicrecords.preload_student_user", views.QueryPatcherPreload("Student.User")).
			WithQueryPatcher("academicrecords.preload_semester", views.QueryPatcherPreload("Semester")).
			WithQueryPatcher("academicrecords.scope_by_role", AcademicRecordScopeByRole),
	)

	// Create view
	lago.RegistryView.Register("academicrecords.CreateView",
		views.CreateView[AcademicRecord](
			lago.GetterRoutePath("academicrecords.DetailRoute", map[string]getters.Getter[any]{
				"id": getters.GetterAny(getters.GetterKey[uint]("$id")),
			}),
		)(
			lago.GetPageView("academicrecords.AcademicRecordCreateForm"),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware),
	)

	// Update view
	lago.RegistryView.Register("academicrecords.UpdateView",
		views.DetailView[AcademicRecord]("academicrecord")(
			views.UpdateView[AcademicRecord](
				lago.GetterRoutePath("academicrecords.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("$id")),
				}),
			)(
				lago.GetPageView("academicrecords.AcademicRecordUpdateForm"),
			),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithQueryPatcher("academicrecords.preload_student_user", views.QueryPatcherPreload("Student.User")).
			WithQueryPatcher("academicrecords.preload_semester", views.QueryPatcherPreload("Semester")).
			WithQueryPatcher("academicrecords.scope_by_role", AcademicRecordScopeByRole),
	)

	// Delete view
	lago.RegistryView.Register("academicrecords.DeleteView",
		views.DetailView[AcademicRecord]("academicrecord")(
			views.DeleteView[AcademicRecord](
				lago.GetterRoutePath("academicrecords.DefaultRoute", nil),
			)(
				lago.GetPageView("academicrecords.AcademicRecordDeleteForm"),
			),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithQueryPatcher("academicrecords.preload_student_user", views.QueryPatcherPreload("Student.User")).
			WithQueryPatcher("academicrecords.preload_semester", views.QueryPatcherPreload("Semester")).
			WithQueryPatcher("academicrecords.scope_by_role", AcademicRecordScopeByRole),
	)

	// Selection view
	lago.RegistryView.Register("academicrecords.SelectView",
		views.ListView[AcademicRecord]("academicrecords")(
			lago.GetPageView("academicrecords.AcademicRecordSelectionTable"),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithQueryPatcher("academicrecords.preload_student_user", views.QueryPatcherPreload("Student.User")).
			WithQueryPatcher("academicrecords.preload_semester", views.QueryPatcherPreload("Semester")).
			WithQueryPatcher("academicrecords.scope_by_role", AcademicRecordScopeByRole),
	)
}
