package p_seer_gdelt

import (
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/views"
)

var gdeltEventListPatchers = views.QueryPatchers[Event]{
	{Key: "seer_gdelt.event_list.order", Value: views.QueryPatcherOrderBy[Event]{Order: "id DESC"}},
}

func init() {
	lago.RegistryView.Register("seer_gdelt.MapView",
		lago.GetPageView("seer_gdelt.MapPage").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_gdelt.map", gdeltMapLayer{}))

	lago.RegistryView.Register("seer_gdelt.SearchView",
		lago.GetPageView("seer_gdelt.SearchPage").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_gdelt.search", gdeltSearchLayer{}))

	lago.RegistryView.Register("seer_gdelt.EventListView",
		lago.GetPageView("seer_gdelt.EventTable").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_gdelt.event.list", views.LayerList[Event]{
				Key:           getters.Static("gdeltEvents"),
				PageSize:      getters.Static(uint(25)),
				QueryPatchers: gdeltEventListPatchers,
			}))

	lago.RegistryView.Register("seer_gdelt.EventCreateView",
		lago.GetPageView("seer_gdelt.EventCreateForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_gdelt.event.create", views.LayerCreate[Event]{
				SuccessURL: lago.RoutePath("seer_gdelt.EventDetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("$id")),
				}),
			}))

	lago.RegistryView.Register("seer_gdelt.EventDetailView",
		lago.GetPageView("seer_gdelt.EventDetail").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_gdelt.event.detail", views.LayerDetail[Event]{
				Key:          getters.Static("gdeltEvent"),
				PathParamKey: getters.Static("id"),
			}))

	lago.RegistryView.Register("seer_gdelt.EventUpdateView",
		lago.GetPageView("seer_gdelt.EventUpdateForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_gdelt.event.detail_for_update", views.LayerDetail[Event]{
				Key:          getters.Static("gdeltEvent"),
				PathParamKey: getters.Static("id"),
			}).
			WithLayer("seer_gdelt.event.update", views.LayerUpdate[Event]{
				Key: getters.Static("gdeltEvent"),
				SuccessURL: lago.RoutePath("seer_gdelt.EventDetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("gdeltEvent.ID")),
				}),
			}))

	lago.RegistryView.Register("seer_gdelt.EventDeleteView",
		lago.GetPageView("seer_gdelt.EventDeleteForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_gdelt.event.detail_for_delete", views.LayerDetail[Event]{
				Key:          getters.Static("gdeltEvent"),
				PathParamKey: getters.Static("id"),
			}).
			WithLayer("seer_gdelt.event.delete", views.LayerDelete[Event]{
				Key:        getters.Static("gdeltEvent"),
				SuccessURL: lago.RoutePath("seer_gdelt.EventListRoute", nil),
			}))
}
