package p_teachers

import (
	"github.com/lariv-in/getters"
	"github.com/lariv-in/lago"
	"github.com/lariv-in/p_users"
	"github.com/lariv-in/views"
)

func init() {
	// List view - displays all teachers with filtering
	lago.RegistryView.Register("teachers.ListView",
		views.ListView[Teacher]("teachers")(
			lago.GetPageView("teachers.TeacherTable")).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithQueryPatcher("teachers.preload", views.QueryPatcherPreload("User")).
			WithQueryPatcher("teachers.order", views.QueryPatcherOrderBy("code ASC")))

	// Detail view - displays a single teacher
	lago.RegistryView.Register("teachers.DetailView",
		views.DetailView[Teacher]("teacher")(
			lago.GetPageView("teachers.TeacherDetail")).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithQueryPatcher("teachers.preload", views.QueryPatcherPreload("User")))

	// Create view - handles creating a new teacher
	lago.RegistryView.Register("teachers.CreateView",
		views.CreateView[Teacher](lago.GetterRoutePath("teachers.DetailRoute", map[string]getters.Getter[any]{"id": getters.GetterAny(getters.GetterKey[uint]("$id"))}))(
			lago.GetPageView("teachers.TeacherCreateForm")).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware))

	// Update view - handles updating an existing teacher
	lago.RegistryView.Register("teachers.UpdateView",
		views.DetailView[Teacher]("teacher")(
			views.UpdateView[Teacher](lago.GetterRoutePath("teachers.DetailRoute", map[string]getters.Getter[any]{"id": getters.GetterAny(getters.GetterKey[uint]("$id"))}))(
				lago.GetPageView("teachers.TeacherUpdateForm"))).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithQueryPatcher("teachers.preload", views.QueryPatcherPreload("User")))

	// Delete view - handles deleting a teacher
	lago.RegistryView.Register("teachers.DeleteView",
		views.DetailView[Teacher]("teacher")(
			views.DeleteView[Teacher](lago.GetterRoutePath("teachers.DefaultRoute", nil))(
				lago.GetPageView("teachers.TeacherDeleteForm"))).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithQueryPatcher("teachers.preload", views.QueryPatcherPreload("User")))

	// Select view - modal table for selecting a teacher (for foreign key inputs)
	lago.RegistryView.Register("teachers.SelectView",
		views.ListView[Teacher]("teachers")(
			lago.GetPageView("teachers.TeacherSelectionTable")).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithQueryPatcher("teachers.preload", views.QueryPatcherPreload("User")).
			WithQueryPatcher("teachers.order", views.QueryPatcherOrderBy("code ASC")))

	// Multi-select view - modal table for selecting multiple teachers
	lago.RegistryView.Register("teachers.MultiSelectView",
		views.ListView[Teacher]("teachers")(
			lago.GetPageView("teachers.TeacherMultiSelectionTable")).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithQueryPatcher("teachers.preload", views.QueryPatcherPreload("User")).
			WithQueryPatcher("teachers.order", views.QueryPatcherOrderBy("code ASC")))
}
