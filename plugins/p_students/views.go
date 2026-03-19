package p_students

import (
	"context"
	"fmt"
	"net/http"

	"github.com/lariv-in/getters"
	"github.com/lariv-in/lago"
	"github.com/lariv-in/p_users"
	"github.com/lariv-in/views"
	"gorm.io/gorm"
)

// uintToStringGetter converts a uint getter to a string getter.
// This is needed for route path building where path params must be strings.
func uintToStringGetter(g getters.Getter[uint]) getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		v, err := g(ctx)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%d", v), nil
	}
}

func init() {
	// List view - displays all students with filtering
	lago.RegistryView.Register("students.ListView",
		views.ListView[Student]("students")(
			lago.GetPageView("students.StudentTable")).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithQueryPatcher("students.preload", preloadUserQuery))

	// Detail view - displays a single student
	lago.RegistryView.Register("students.DetailView",
		views.DetailView[Student]("student")(
			lago.GetPageView("students.StudentDetail")).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithQueryPatcher("students.preload", preloadUserQuery))

	// Create view - handles creating a new student
	lago.RegistryView.Register("students.CreateView",
		views.CreateView[Student](lago.GetterRoutePath("students.DetailRoute", map[string]getters.Getter[any]{"id": getters.GetterAny(uintToStringGetter(getters.GetterKey[uint]("$id")))}))(
			lago.GetPageView("students.StudentCreateForm")).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware))

	// Update view - handles updating an existing student
	lago.RegistryView.Register("students.UpdateView",
		views.DetailView[Student]("student")(
			views.UpdateView[Student](lago.GetterRoutePath("students.DetailRoute", map[string]getters.Getter[any]{"id": getters.GetterAny(uintToStringGetter(getters.GetterKey[uint]("$id")))}))(
				lago.GetPageView("students.StudentUpdateForm"))).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithQueryPatcher("students.preload", preloadUserQuery))

	// Delete view - handles deleting a student
	lago.RegistryView.Register("students.DeleteView",
		views.DetailView[Student]("student")(
			views.DeleteView[Student](lago.GetterRoutePath("students.DefaultRoute", nil))(
				lago.GetPageView("students.StudentDeleteForm"))).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithQueryPatcher("students.preload", preloadUserQuery))

	// Select view - modal table for selecting a student (for foreign key inputs)
	lago.RegistryView.Register("students.SelectView",
		views.ListView[Student]("students")(
			lago.GetPageView("students.StudentSelectionTable")).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithQueryPatcher("students.preload", preloadUserQuery))
}

// preloadUserQuery ensures the User relation is loaded for display purposes.
// This is necessary because the table and detail pages need to show User.Name and User.Email.
func preloadUserQuery(v *views.View, r *http.Request, query *gorm.DB) *gorm.DB {
	return query.Preload("User")
}
