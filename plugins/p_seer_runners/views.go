package p_seer_runners

import (
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/views"
)

func init() {
	runnerListPatchers := views.QueryPatchers[Runner]{
		{Key: "seer_runners.runner.order", Value: views.QueryPatcherOrderBy[Runner]{Order: "id DESC"}},
	}

	lago.RegistryView.Register("seer_runners.ListView",
		lago.GetPageView("seer_runners.RunnerTable").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_runners.runner.list", views.LayerList[Runner]{
				Key:           getters.Static("runners"),
				QueryPatchers: runnerListPatchers,
			}))

	lago.RegistryView.Register("seer_runners.CreateView",
		lago.GetPageView("seer_runners.RunnerCreateForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_runners.runner.create", views.LayerCreate[Runner]{
				SuccessURL: lago.RoutePath("seer_runners.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("$id")),
				}),
			}))

	lago.RegistryView.Register("seer_runners.DetailView",
		lago.GetPageView("seer_runners.RunnerDetail").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_runners.runner.detail", views.LayerDetail[Runner]{
				Key:          getters.Static("runner"),
				PathParamKey: getters.Static("id"),
			}))
}
