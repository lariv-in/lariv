package p_teachers

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
	// List view - displays all teachers with filtering
	lago.RegistryView.Register("teachers.ListView",
		views.ListView[Teacher]("teachers")(
			lago.GetPageView("teachers.TeacherTable")).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithQueryPatcher("teachers.preload", preloadUserQuery).
			WithQueryPatcher("teachers.order", orderByCodeQuery))

	// Detail view - displays a single teacher
	lago.RegistryView.Register("teachers.DetailView",
		views.DetailView[Teacher]("teacher")(
			lago.GetPageView("teachers.TeacherDetail")).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithQueryPatcher("teachers.preload", preloadUserQuery))

	// Create view - handles creating a new teacher
	lago.RegistryView.Register("teachers.CreateView",
		views.CreateView[Teacher](lago.GetterRoutePath("teachers.DetailRoute", map[string]getters.Getter[any]{"id": getters.GetterAny(uintToStringGetter(getters.GetterKey[uint]("$id")))}))(
			lago.GetPageView("teachers.TeacherCreateForm")).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware))

	// Update view - handles updating an existing teacher
	lago.RegistryView.Register("teachers.UpdateView",
		views.DetailView[Teacher]("teacher")(
			views.UpdateView[Teacher](lago.GetterRoutePath("teachers.DetailRoute", map[string]getters.Getter[any]{"id": getters.GetterAny(uintToStringGetter(getters.GetterKey[uint]("$id")))}))(
				lago.GetPageView("teachers.TeacherUpdateForm"))).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithQueryPatcher("teachers.preload", preloadUserQuery))

	// Delete view - handles deleting a teacher
	lago.RegistryView.Register("teachers.DeleteView",
		views.DetailView[Teacher]("teacher")(
			views.DeleteView[Teacher](lago.GetterRoutePath("teachers.DefaultRoute", nil))(
				lago.GetPageView("teachers.TeacherDeleteForm"))).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithQueryPatcher("teachers.preload", preloadUserQuery))

	// Select view - modal table for selecting a teacher (for foreign key inputs)
	lago.RegistryView.Register("teachers.SelectView",
		views.ListView[Teacher]("teachers")(
			lago.GetPageView("teachers.TeacherSelectionTable")).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithQueryPatcher("teachers.preload", preloadUserQuery).
			WithQueryPatcher("teachers.order", orderByCodeQuery))
}

// preloadUserQuery ensures the User relation is loaded for display purposes.
// This is necessary because the table and detail pages need to show User.Name and User.Email.
func preloadUserQuery(v *views.View, r *http.Request, query *gorm.DB) *gorm.DB {
	return query.Preload("User")
}

// orderByCodeQuery applies the default ordering by Code (matching Django's ordering = ["code"]).
func orderByCodeQuery(v *views.View, r *http.Request, query *gorm.DB) *gorm.DB {
	return query.Order("code ASC")
}
