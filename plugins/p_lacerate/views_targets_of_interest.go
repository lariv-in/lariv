package p_lacerate

import (
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/views"
)

func init() {
	patchers := views.QueryPatchers[TargetOfInterest]{
		{Key: "lacerate.targets_of_interest.order_id", Value: views.QueryPatcherOrderBy[TargetOfInterest]{Order: "id DESC"}},
	}

	lago.RegistryView.Register("lacerate.TargetOfInterestListView",
		lago.GetPageView("lacerate.TargetsOfInterestTable").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("lacerate.targets_of_interest.list", views.LayerList[TargetOfInterest]{
				Key:           getters.Static("targets_of_interest"),
				QueryPatchers: patchers,
			}))

	lago.RegistryView.Register("lacerate.TargetOfInterestDetailView",
		lago.GetPageView("lacerate.TargetOfInterestDetail").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("lacerate.targets_of_interest.detail", views.LayerDetail[TargetOfInterest]{
				Key:          getters.Static("target_of_interest"),
				PathParamKey: getters.Static("id"),
			}))

	lago.RegistryView.Register("lacerate.TargetOfInterestCreateView",
		lago.GetPageView("lacerate.TargetOfInterestCreateForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("lacerate.targets_of_interest.create", views.LayerCreate[TargetOfInterest]{
				SuccessURL: lago.RoutePath("lacerate.TargetOfInterestDetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("$id")),
				}),
			}))

	lago.RegistryView.Register("lacerate.TargetOfInterestUpdateView",
		lago.GetPageView("lacerate.TargetOfInterestUpdateForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("lacerate.targets_of_interest.update_detail", views.LayerDetail[TargetOfInterest]{
				Key:          getters.Static("target_of_interest"),
				PathParamKey: getters.Static("id"),
			}).
			WithLayer("lacerate.targets_of_interest.update", views.LayerUpdate[TargetOfInterest]{
				Key: getters.Static("target_of_interest"),
				SuccessURL: lago.RoutePath("lacerate.TargetOfInterestDetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("target_of_interest.ID")),
				}),
			}))

	lago.RegistryView.Register("lacerate.TargetOfInterestDeleteView",
		lago.GetPageView("lacerate.TargetOfInterestDeleteForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("lacerate.targets_of_interest.delete_detail", views.LayerDetail[TargetOfInterest]{
				Key:          getters.Static("target_of_interest"),
				PathParamKey: getters.Static("id"),
			}).
			WithLayer("lacerate.targets_of_interest.delete", views.LayerDelete[TargetOfInterest]{
				Key:        getters.Static("target_of_interest"),
				SuccessURL: lago.RoutePath("lacerate.TargetOfInterestListRoute", nil),
			}))
}
