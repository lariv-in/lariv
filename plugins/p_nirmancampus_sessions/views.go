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

func (sessionIsActiveFilterQueryPatcher) Patch(_ views.View, r *http.Request, query gorm.ChainInterface[AdmissionSession]) gorm.ChainInterface[AdmissionSession] {
	return applyIsActiveFilterToQuery(r, query)
}

func applyIsActiveFilterToQuery[R any](r *http.Request, query gorm.ChainInterface[R]) gorm.ChainInterface[R] {
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
	listPageSize := getters.Static[uint](200)

	// List: admission sessions ("All Sessions").
	lago.RegistryView.Register("sessions.ListView",
		lago.GetPageView("sessions.SessionTable").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("sessions.list", views.LayerList[AdmissionSession]{
				Key:      getters.Static("sessions"),
				PageSize: listPageSize,
				QueryPatchers: views.QueryPatchers[AdmissionSession]{
					registry.Pair[string, views.QueryPatcher[AdmissionSession]]{
						Key:   "sessions.filter_is_active",
						Value: sessionIsActiveFilterQueryPatcher{},
					},
				},
			}))

	// Admission detail
	lago.RegistryView.Register("sessions.DetailView",
		lago.GetPageView("sessions.SessionDetail").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("sessions.detail", views.LayerDetail[AdmissionSession]{
				Key:          getters.Static("session"),
				PathParamKey: getters.Static("id"),
			}))

	// Admission create
	lago.RegistryView.Register("sessions.CreateView",
		lago.GetPageView("sessions.SessionCreateForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("sessions.create", views.LayerCreate[AdmissionSession]{
				SuccessURL: lago.RoutePath("sessions.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("$id")),
				}),
			}))

	// Admission update
	lago.RegistryView.Register("sessions.UpdateView",
		lago.GetPageView("sessions.SessionUpdateForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("sessions.detail", views.LayerDetail[AdmissionSession]{
				Key:          getters.Static("session"),
				PathParamKey: getters.Static("id"),
			}).
			WithLayer("sessions.update", views.LayerUpdate[AdmissionSession]{
				Key: getters.Static("session"),
				SuccessURL: lago.RoutePath("sessions.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("session.ID")),
				}),
			}))

	// Admission delete
	lago.RegistryView.Register("sessions.DeleteView",
		lago.GetPageView("sessions.SessionDeleteForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("sessions.detail", views.LayerDetail[AdmissionSession]{
				Key:          getters.Static("session"),
				PathParamKey: getters.Static("id"),
			}).
			WithLayer("sessions.delete", views.LayerDelete[AdmissionSession]{
				Key:        getters.Static("session"),
				SuccessURL: lago.RoutePath("sessions.DefaultRoute", nil),
			}))

	// Admission selection (e.g. academic record FK)
	lago.RegistryView.Register("sessions.SelectView",
		lago.GetPageView("sessions.sessionselectionTable").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("sessions.select", views.LayerList[AdmissionSession]{
				Key:      getters.Static("sessions"),
				PageSize: listPageSize,
				QueryPatchers: views.QueryPatchers[AdmissionSession]{
					registry.Pair[string, views.QueryPatcher[AdmissionSession]]{
						Key:   "sessions.filter_is_active",
						Value: sessionIsActiveFilterQueryPatcher{},
					},
				},
			}))
}
