package p_students

import (
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/p_users"
	"github.com/lariv-in/lago/views"
)

func init() {
	// List view - displays all students with filtering
	lago.RegistryView.Register("students.ListView",
		views.ListView[Student]("students")(
			lago.GetPageView("students.StudentTable")).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithQueryPatcher("students.preload", views.QueryPatcherPreload("User")))

	// Detail view - displays a single student
	lago.RegistryView.Register("students.DetailView",
		views.DetailView[Student]("student")(
			lago.GetPageView("students.StudentDetail")).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithQueryPatcher("students.preload_user", views.QueryPatcherPreload("User")).
			WithQueryPatcher("students.preload_assets", views.QueryPatcherPreload("Assets")))

	// Create view - handles creating a new student
	lago.RegistryView.Register("students.CreateView",
		views.CreateView[Student](lago.GetterRoutePath("students.DetailRoute", map[string]getters.Getter[any]{"id": getters.GetterAny(getters.GetterKey[uint]("$id"))}))(
			lago.GetPageView("students.StudentCreateForm")).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware))

	// Update view - handles updating an existing student
	lago.RegistryView.Register("students.UpdateView",
		views.DetailView[Student]("student")(
			views.UpdateView[Student](lago.GetterRoutePath("students.DetailRoute", map[string]getters.Getter[any]{"id": getters.GetterAny(getters.GetterKey[uint]("$id"))}))(
				lago.GetPageView("students.StudentUpdateForm"))).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithQueryPatcher("students.preload_user", views.QueryPatcherPreload("User")).
			WithQueryPatcher("students.preload_assets", views.QueryPatcherPreload("Assets")))

	// Delete view - handles deleting a student
	lago.RegistryView.Register("students.DeleteView",
		views.DetailView[Student]("student")(
			views.DeleteView[Student](lago.GetterRoutePath("students.DefaultRoute", nil))(
				lago.GetPageView("students.StudentDeleteForm"))).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithQueryPatcher("students.preload_user", views.QueryPatcherPreload("User")).
			WithQueryPatcher("students.preload_assets", views.QueryPatcherPreload("Assets")))

	// Select view - modal table for selecting a student (for foreign key inputs)
	lago.RegistryView.Register("students.SelectView",
		views.ListView[Student]("students")(
			lago.GetPageView("students.StudentSelectionTable")).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithQueryPatcher("students.preload", views.QueryPatcherPreload("User")))
}
