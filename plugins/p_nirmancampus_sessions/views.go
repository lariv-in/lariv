package p_nirmancampus_sessions

import (
	"net/http"

	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/views"
	"gorm.io/gorm"
)

func init() {
	// List view.
	lago.RegistryView.Register("sessions.ListView",
		views.ListView[Semester]("sessions")(
			lago.GetPageView("sessions.SemesterTable"),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithQueryPatcher("sessions.filter_is_active", func(_ *views.View, r *http.Request, query *gorm.DB) *gorm.DB {
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
	lago.RegistryView.Register("sessions.DetailView",
		views.DetailView[Semester]("semester", "id")(
			lago.GetPageView("sessions.SemesterDetail"),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware))

	// Create view.
	lago.RegistryView.Register("sessions.CreateView",
		views.CreateView[Semester](
			lago.RoutePath("sessions.DetailRoute", map[string]getters.Getter[any]{
				"id": getters.Any(getters.Key[uint]("$id")),
			}),
		)(
			lago.GetPageView("sessions.SemesterCreateForm"),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware))

	// Update view.
	lago.RegistryView.Register("sessions.UpdateView",
		views.DetailView[Semester]("semester", "id")(
			views.UpdateView[Semester]("id",
				lago.RoutePath("sessions.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("$id")),
				}),
			)(
				lago.GetPageView("sessions.SemesterUpdateForm"),
			),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware))

	// Delete view.
	lago.RegistryView.Register("sessions.DeleteView",
		views.DetailView[Semester]("semester", "id")(
			views.DeleteView[Semester]("id",
				lago.RoutePath("sessions.DefaultRoute", nil),
			)(
				lago.GetPageView("sessions.SemesterDeleteForm"),
			),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware))

	// Selection view.
	lago.RegistryView.Register("sessions.SelectView",
		views.ListView[Semester]("sessions")(
			lago.GetPageView("sessions.sessionselectionTable"),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithQueryPatcher("sessions.filter_is_active", func(_ *views.View, r *http.Request, query *gorm.DB) *gorm.DB {
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
