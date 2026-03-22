package p_totschool_tally

import (
	"github.com/lariv-in/lago/lago"
)

func init() {
	_ = lago.RegistryRoute.Register("tally.TallyListRoute", lago.Route{
		Path:    "/tally/list/",
		Handler: lago.NewDynamicView("tally.TallyListView"),
	})

	_ = lago.RegistryRoute.Register("tally.TallyDashboardRoute", lago.Route{
		Path:    "/tally/",
		Handler: lago.NewDynamicView("tally.TallyDashboardView"),
	})

	_ = lago.RegistryRoute.Register("tally.TallyLeaderboardRoute", lago.Route{
		Path:    "/tally/leaderboard/",
		Handler: lago.NewDynamicView("tally.TallyLeaderboardView"),
	})

	_ = lago.RegistryRoute.Register("tally.TallyDailyFormRoute", lago.Route{
		Path:    "/tally/daily/",
		Handler: lago.NewDynamicView("tally.TallyDailyFormView"),
	})

	_ = lago.RegistryRoute.Register("tally.TallyCreateRoute", lago.Route{
		Path:    "/tally/create/",
		Handler: lago.NewDynamicView("tally.TallyCreateView"),
	})

	_ = lago.RegistryRoute.Register("tally.TallyUpdateRoute", lago.Route{
		Path:    "/tally/{id}/update/",
		Handler: lago.NewDynamicView("tally.TallyUpdateView"),
	})

	_ = lago.RegistryRoute.Register("tally.TallyDeleteRoute", lago.Route{
		Path:    "/tally/{id}/delete/",
		Handler: lago.NewDynamicView("tally.TallyDeleteView"),
	})

	_ = lago.RegistryRoute.Register("tally.TallyDetailRoute", lago.Route{
		Path:    "/tally/{id}/",
		Handler: lago.NewDynamicView("tally.TallyDetailView"),
	})
}
