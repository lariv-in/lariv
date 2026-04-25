package p_seer_opensky

import (
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/views"
)

var stateListQueryPatchers = views.QueryPatchers[OpenSkyState]{
	{Key: "seer_opensky.state_list.order", Value: views.QueryPatcherOrderBy[OpenSkyState]{Order: "id DESC"}},
}

func init() {
	lago.RegistryView.Register("seer_opensky.MapView",
		lago.GetPageView("seer_opensky.MapPage").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_opensky.map", openSkyMapLayer{}))

	lago.RegistryView.Register("seer_opensky.StateListView",
		lago.GetPageView("seer_opensky.StateTablePage").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_opensky.state.list", views.LayerList[OpenSkyState]{
				Key:           getters.Static("openskyStates"),
				PageSize:      getters.Static(uint(25)),
				QueryPatchers: stateListQueryPatchers,
			}))

		lago.RegistryView.Register("seer_opensky.StateCreateView",
		lago.GetPageView("seer_opensky.StateCreateForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_opensky.state.create", views.LayerCreate[OpenSkyState]{
				SuccessURL: lago.RoutePath("seer_opensky.StateDetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("$id")),
				}),
				FormPatchers: openSkyFormPatchers,
			}))

	lago.RegistryView.Register("seer_opensky.StateDetailView",
		lago.GetPageView("seer_opensky.StateDetailPage").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_opensky.state.detail", views.LayerDetail[OpenSkyState]{
				Key:          getters.Static("openskyState"),
				PathParamKey: getters.Static("id"),
			}))

	lago.RegistryView.Register("seer_opensky.StateUpdateView",
		lago.GetPageView("seer_opensky.StateUpdateForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_opensky.state.detail_for_update", views.LayerDetail[OpenSkyState]{
				Key:          getters.Static("openskyState"),
				PathParamKey: getters.Static("id"),
			}).
			WithLayer("seer_opensky.state.update", views.LayerUpdate[OpenSkyState]{
				Key: getters.Static("openskyState"),
				SuccessURL: lago.RoutePath("seer_opensky.StateDetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("openskyState.ID")),
				}),
				FormPatchers: openSkyFormPatchers,
			}))

	lago.RegistryView.Register("seer_opensky.StateDeleteView",
		lago.GetPageView("seer_opensky.StateDeleteFormModal").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_opensky.state.detail_for_delete", views.LayerDetail[OpenSkyState]{
				Key:          getters.Static("openskyState"),
				PathParamKey: getters.Static("id"),
			}).
			WithLayer("seer_opensky.state.delete", views.LayerDelete[OpenSkyState]{
				Key:        getters.Static("openskyState"),
				SuccessURL: lago.RoutePath("seer_opensky.StateListRoute", nil),
			}))
}
