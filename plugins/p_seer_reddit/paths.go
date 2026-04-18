package p_seer_reddit

import (
	"github.com/lariv-in/lago/lago"
)

func registerRoutes() {
	_ = lago.RegistryRoute.Register("seer_reddit.DefaultRoute", lago.Route{
		Path:    AppUrl,
		Handler: lago.NewDynamicView("seer_reddit.RedditSourceListView"),
	})

	_ = lago.RegistryRoute.Register("seer_reddit.RedditSourceDetailRoute", lago.Route{
		Path:    AppUrl + "sources/{id}/",
		Handler: lago.NewDynamicView("seer_reddit.RedditSourceDetailView"),
	})

	_ = lago.RegistryRoute.Register("seer_reddit.RedditPostListRoute", lago.Route{
		Path:    AppUrl + "posts/",
		Handler: lago.NewDynamicView("seer_reddit.RedditPostListView"),
	})

	_ = lago.RegistryRoute.Register("seer_reddit.RedditPostDetailRoute", lago.Route{
		Path:    AppUrl + "posts/{id}/",
		Handler: lago.NewDynamicView("seer_reddit.RedditPostDetailView"),
	})
}

func init() {
	registerRoutes()
}
