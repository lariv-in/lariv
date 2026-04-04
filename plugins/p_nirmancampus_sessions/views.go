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

type sessionIsActiveFilterQueryPatcher struct{}

func (sessionIsActiveFilterQueryPatcher) Patch(_ views.View, r *http.Request, query gorm.ChainInterface[Session]) gorm.ChainInterface[Session] {
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
		lago.GetPageView("sessions.SessionTable").
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware{}).
			WithMiddleware("sessions.list", views.MiddlewareList[Session]{
				Key: getters.Static("sessions"),
				QueryPatchers: views.QueryPatchers[Session]{
					registry.Pair[string, views.QueryPatcher[Session]]{
						Key:   "sessions.filter_is_active",
						Value: sessionIsActiveFilterQueryPatcher{},
					},
				},
			}))

	// Detail view.
	lago.RegistryView.Register("sessions.DetailView",
		lago.GetPageView("sessions.SessionDetail").
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware{}).
			WithMiddleware("sessions.detail", views.MiddlewareDetail[Session]{
				Key:          getters.Static("session"),
				PathParamKey: getters.Static("id"),
			}))

	// Create view.
	lago.RegistryView.Register("sessions.CreateView",
		lago.GetPageView("sessions.SessionCreateForm").
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware{}).
			WithMiddleware("sessions.create", views.MiddlewareCreate[Session]{
				SuccessURL: lago.RoutePath("sessions.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("$id")),
				}),
			}))

	// Update view.
	lago.RegistryView.Register("sessions.UpdateView",
		lago.GetPageView("sessions.SessionUpdateForm").
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware{}).
			WithMiddleware("sessions.detail", views.MiddlewareDetail[Session]{
				Key:          getters.Static("session"),
				PathParamKey: getters.Static("id"),
			}).
			WithMiddleware("sessions.update", views.MiddlewareUpdate[Session]{
				Key: getters.Static("session"),
				SuccessURL: lago.RoutePath("sessions.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("session.ID")),
				}),
			}))

	// Delete view.
	lago.RegistryView.Register("sessions.DeleteView",
		lago.GetPageView("sessions.SessionDeleteForm").
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware{}).
			WithMiddleware("sessions.detail", views.MiddlewareDetail[Session]{
				Key:          getters.Static("session"),
				PathParamKey: getters.Static("id"),
			}).
			WithMiddleware("sessions.delete", views.MiddlewareDelete[Session]{
				Key:        getters.Static("session"),
				SuccessURL: lago.RoutePath("sessions.DefaultRoute", nil),
			}))

	// Selection view.
	lago.RegistryView.Register("sessions.SelectView",
		lago.GetPageView("sessions.sessionselectionTable").
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware{}).
			WithMiddleware("sessions.select", views.MiddlewareList[Session]{
				Key: getters.Static("sessions"),
				QueryPatchers: views.QueryPatchers[Session]{
					registry.Pair[string, views.QueryPatcher[Session]]{
						Key:   "sessions.filter_is_active",
						Value: sessionIsActiveFilterQueryPatcher{},
					},
				},
			}))
}
