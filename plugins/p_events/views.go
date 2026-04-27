package p_events

import (
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/registry"
	"github.com/lariv-in/lago/views"
)

var schoolEventPreload = views.QueryPatchers[SchoolEvent]{
	registry.Pair[string, views.QueryPatcher[SchoolEvent]]{Key: "events.fk", Value: views.QueryPatcherPreload[SchoolEvent]{Fields: []string{"Semester"}}},
}

func init() {
	lago.RegistryView.Register("events.ListView",
		lago.GetPageView("events.SchoolEventTable").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("events.list", views.LayerList[SchoolEvent]{Key: getters.Static("school_events")}))

	lago.RegistryView.Register("events.DetailView",
		lago.GetPageView("events.SchoolEventDetail").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("events.detail", views.LayerDetail[SchoolEvent]{
				Key: getters.Static("school_event"), PathParamKey: getters.Static("id"), QueryPatchers: schoolEventPreload,
			}))

	lago.RegistryView.Register("events.CreateView",
		lago.GetPageView("events.SchoolEventCreateForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("events.create", views.LayerCreate[SchoolEvent]{
				SuccessURL: lago.RoutePath("events.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("$id")),
				}),
			}))

	lago.RegistryView.Register("events.UpdateView",
		lago.GetPageView("events.SchoolEventUpdateForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("events.detail", views.LayerDetail[SchoolEvent]{
				Key: getters.Static("school_event"), PathParamKey: getters.Static("id"), QueryPatchers: schoolEventPreload,
			}).
			WithLayer("events.update", views.LayerUpdate[SchoolEvent]{
				Key: getters.Static("school_event"),
				SuccessURL: lago.RoutePath("events.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("school_event.ID")),
				}),
			}))

	lago.RegistryView.Register("events.DeleteView",
		lago.GetPageView("events.SchoolEventDeleteForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("events.detail", views.LayerDetail[SchoolEvent]{
				Key: getters.Static("school_event"), PathParamKey: getters.Static("id"), QueryPatchers: schoolEventPreload,
			}).
			WithLayer("events.delete", views.LayerDelete[SchoolEvent]{
				Key: getters.Static("school_event"), SuccessURL: lago.RoutePath("events.DefaultRoute", nil),
			}))
}
