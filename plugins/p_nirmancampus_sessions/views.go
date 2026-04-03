package p_nirmancampus_sessions

import (
	"net/http"

	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/registry"
	"github.com/lariv-in/lago/views"
	"gorm.io/gorm"
)

type semesterIsActiveFilterQueryPatcher struct{}

func (semesterIsActiveFilterQueryPatcher) Patch(_ views.View, r *http.Request, query gorm.ChainInterface[Semester]) gorm.ChainInterface[Semester] {
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
}

func init() {
	// List view.
	lago.RegistryView.Register("sessions.ListView",
		lago.GetPageView("sessions.SemesterTable").
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware{}).
			WithMiddleware("sessions.list", views.MiddlewareList[Semester]{
				Key: getters.Static("sessions"),
				QueryPatchers: views.QueryPatchers[Semester]{
					registry.Pair[string, views.QueryPatcher[Semester]]{
						Key:   "sessions.filter_is_active",
						Value: semesterIsActiveFilterQueryPatcher{},
					},
				},
			}))

	// Detail view.
	lago.RegistryView.Register("sessions.DetailView",
		lago.GetPageView("sessions.SemesterDetail").
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware{}).
			WithMiddleware("sessions.detail", views.MiddlewareDetail[Semester]{
				Key:          getters.Static("semester"),
				PathParamKey: getters.Static("id"),
			}))

	// Create view.
	lago.RegistryView.Register("sessions.CreateView",
		lago.GetPageView("sessions.SemesterCreateForm").
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware{}).
			WithMiddleware("sessions.create", views.MiddlewareCreate[Semester]{
				SuccessURL: lago.RoutePath("sessions.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("$id")),
				}),
			}))

	// Update view.
	lago.RegistryView.Register("sessions.UpdateView",
		lago.GetPageView("sessions.SemesterUpdateForm").
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware{}).
			WithMiddleware("sessions.detail", views.MiddlewareDetail[Semester]{
				Key:          getters.Static("semester"),
				PathParamKey: getters.Static("id"),
			}).
			WithMiddleware("sessions.update", views.MiddlewareUpdate[Semester]{
				Key: getters.Static("semester"),
				SuccessURL: lago.RoutePath("sessions.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("$id")),
				}),
			}))

	// Delete view.
	lago.RegistryView.Register("sessions.DeleteView",
		lago.GetPageView("sessions.SemesterDeleteForm").
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware{}).
			WithMiddleware("sessions.detail", views.MiddlewareDetail[Semester]{
				Key:          getters.Static("semester"),
				PathParamKey: getters.Static("id"),
			}).
			WithMiddleware("sessions.delete", views.MiddlewareDelete[Semester]{
				Key:        getters.Static("semester"),
				SuccessURL: lago.RoutePath("sessions.DefaultRoute", nil),
			}))

	// Selection view.
	lago.RegistryView.Register("sessions.SelectView",
		lago.GetPageView("sessions.sessionselectionTable").
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware{}).
			WithMiddleware("sessions.select", views.MiddlewareList[Semester]{
				Key: getters.Static("sessions"),
				QueryPatchers: views.QueryPatchers[Semester]{
					registry.Pair[string, views.QueryPatcher[Semester]]{
						Key:   "sessions.filter_is_active",
						Value: semesterIsActiveFilterQueryPatcher{},
					},
				},
			}))
}
