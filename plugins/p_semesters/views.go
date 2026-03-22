package p_semesters

import (
	"net/http"

	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/p_users"
	"github.com/lariv-in/lago/views"
	"gorm.io/gorm"
)

func init() {
	// List view.
	lago.RegistryView.Register("semesters.ListView",
		views.ListView[Semester]("semesters")(
			lago.GetPageView("semesters.SemesterTable"),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithQueryPatcher("semesters.filter_is_active", func(_ *views.View, r *http.Request, query *gorm.DB) *gorm.DB {
				getMap, ok := r.Context().Value("$get").(map[string]any)
				if !ok {
					return query
				}
				raw, ok := getMap["IsActiveFilter"]
				if !ok || raw == nil {
					return query
				}

				switch typed := raw.(type) {
				case bool:
					return query.Where("is_active = ?", typed)
				case string:
					if typed == "True" || typed == "true" {
						return query.Where("is_active = ?", true)
					}
					if typed == "False" || typed == "false" {
						return query.Where("is_active = ?", false)
					}
					return query
				default:
					return query
				}
			}))

	// Detail view.
	lago.RegistryView.Register("semesters.DetailView",
		views.DetailView[Semester]("semester")(
			lago.GetPageView("semesters.SemesterDetail"),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware))

	// Create view.
	lago.RegistryView.Register("semesters.CreateView",
		views.CreateView[Semester](
			lago.GetterRoutePath("semesters.DetailRoute", map[string]getters.Getter[any]{
				"id": getters.GetterAny(getters.GetterKey[uint]("$id")),
			}),
		)(
			lago.GetPageView("semesters.SemesterCreateForm"),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware))

	// Update view.
	lago.RegistryView.Register("semesters.UpdateView",
		views.DetailView[Semester]("semester")(
			views.UpdateView[Semester](
				lago.GetterRoutePath("semesters.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("$id")),
				}),
			)(
				lago.GetPageView("semesters.SemesterUpdateForm"),
			),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware))

	// Delete view.
	lago.RegistryView.Register("semesters.DeleteView",
		views.DetailView[Semester]("semester")(
			views.DeleteView[Semester](
				lago.GetterRoutePath("semesters.DefaultRoute", nil),
			)(
				lago.GetPageView("semesters.SemesterDeleteForm"),
			),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware))

	// Selection view.
	lago.RegistryView.Register("semesters.SelectView",
		views.ListView[Semester]("semesters")(
			lago.GetPageView("semesters.SemesterSelectionTable"),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithQueryPatcher("semesters.filter_is_active", func(_ *views.View, r *http.Request, query *gorm.DB) *gorm.DB {
				getMap, ok := r.Context().Value("$get").(map[string]any)
				if !ok {
					return query
				}
				raw, ok := getMap["IsActiveFilter"]
				if !ok || raw == nil {
					return query
				}

				switch typed := raw.(type) {
				case bool:
					return query.Where("is_active = ?", typed)
				case string:
					if typed == "True" || typed == "true" {
						return query.Where("is_active = ?", true)
					}
					if typed == "False" || typed == "false" {
						return query.Where("is_active = ?", false)
					}
					return query
				default:
					return query
				}
			}))
}

