package p_seer_reddit

import (
	"github.com/lariv-in/lago/lago"
)

func registerRoutes() {
	_ = lago.RegistryRoute.Register("seer_reddit.DefaultRoute", lago.Route{
		Path:    AppUrl,
		Handler: lago.NewDynamicView("seer_reddit.RedditSourceListView"),
	})

	_ = lago.RegistryRoute.Register("seer_reddit.RedditSourceCreateRoute", lago.Route{
		Path:    AppUrl + "sources/create/",
		Handler: lago.NewDynamicView("seer_reddit.RedditSourceCreateView"),
	})

	_ = lago.RegistryRoute.Register("seer_reddit.RedditPostListBySourceRoute", lago.Route{
		Path:    AppUrl + "sources/{source_id}/posts/",
		Handler: lago.NewDynamicView("seer_reddit.RedditPostListBySourceView"),
	})

	_ = lago.RegistryRoute.Register("seer_reddit.RedditPostListBySourceBulkAddIntelRoute", lago.Route{
		Path:    AppUrl + "sources/{source_id}/posts/bulk-add-intel/",
		Handler: lago.NewDynamicView("seer_reddit.RedditPostListBySourceBulkAddIntelView"),
	})

	_ = lago.RegistryRoute.Register("seer_reddit.RedditSourceFetchPostsRoute", lago.Route{
		Path:    AppUrl + "sources/{source_id}/fetch-posts/",
		Handler: lago.NewDynamicView("seer_reddit.RedditSourceFetchPostsView"),
	})

	_ = lago.RegistryRoute.Register("seer_reddit.RedditSourceDetailRoute", lago.Route{
		Path:    AppUrl + "sources/{id}/",
		Handler: lago.NewDynamicView("seer_reddit.RedditSourceDetailView"),
	})

	_ = lago.RegistryRoute.Register("seer_reddit.RedditSourceUpdateRoute", lago.Route{
		Path:    AppUrl + "sources/{id}/edit/",
		Handler: lago.NewDynamicView("seer_reddit.RedditSourceUpdateView"),
	})

	_ = lago.RegistryRoute.Register("seer_reddit.RedditSourceDeleteRoute", lago.Route{
		Path:    AppUrl + "sources/{id}/delete/",
		Handler: lago.NewDynamicView("seer_reddit.RedditSourceDeleteView"),
	})

	_ = lago.RegistryRoute.Register("seer_reddit.RedditPostListRoute", lago.Route{
		Path:    AppUrl + "posts/",
		Handler: lago.NewDynamicView("seer_reddit.RedditPostListView"),
	})

	_ = lago.RegistryRoute.Register("seer_reddit.RedditPostListBulkAddIntelRoute", lago.Route{
		Path:    AppUrl + "posts/bulk-add-intel/",
		Handler: lago.NewDynamicView("seer_reddit.RedditPostListBulkAddIntelView"),
	})

	_ = lago.RegistryRoute.Register("seer_reddit.RedditPostDetailRoute", lago.Route{
		Path:    AppUrl + "posts/{id}/",
		Handler: lago.NewDynamicView("seer_reddit.RedditPostDetailView"),
	})

	_ = lago.RegistryRoute.Register("seer_reddit.RedditPostDeleteRoute", lago.Route{
		Path:    AppUrl + "posts/{id}/delete/",
		Handler: lago.NewDynamicView("seer_reddit.RedditPostSoftDeleteView"),
	})

	_ = lago.RegistryRoute.Register("seer_reddit.RedditPostAddIntelRoute", lago.Route{
		Path:    AppUrl + "posts/{id}/add-intel/",
		Handler: lago.NewDynamicView("seer_reddit.RedditPostAddIntelView"),
	})

	_ = lago.RegistryRoute.Register("seer_reddit.RedditRunnerListRoute", lago.Route{
		Path:    AppUrl + "workers/",
		Handler: lago.NewDynamicView("seer_reddit.RedditRunnerListView"),
	})

	_ = lago.RegistryRoute.Register("seer_reddit.RedditRunnerCreateRoute", lago.Route{
		Path:    AppUrl + "workers/create/",
		Handler: lago.NewDynamicView("seer_reddit.RedditRunnerCreateView"),
	})

	_ = lago.RegistryRoute.Register("seer_reddit.RedditRunnerSelectRoute", lago.Route{
		Path:    AppUrl + "workers/select/",
		Handler: lago.NewDynamicView("seer_reddit.RedditRunnerSelectView"),
	})

	_ = lago.RegistryRoute.Register("seer_reddit.RedditRunnerWorkerPoolStartRoute", lago.Route{
		Path:    AppUrl + "workers/{id}/worker-pool/start/",
		Handler: lago.NewDynamicView("seer_reddit.RedditRunnerWorkerPoolStartView"),
	})

	_ = lago.RegistryRoute.Register("seer_reddit.RedditRunnerWorkerPoolStopRoute", lago.Route{
		Path:    AppUrl + "workers/{id}/worker-pool/stop/",
		Handler: lago.NewDynamicView("seer_reddit.RedditRunnerWorkerPoolStopView"),
	})

	_ = lago.RegistryRoute.Register("seer_reddit.RedditRunnerDetailRoute", lago.Route{
		Path:    AppUrl + "workers/{id}/",
		Handler: lago.NewDynamicView("seer_reddit.RedditRunnerDetailView"),
	})

	_ = lago.RegistryRoute.Register("seer_reddit.RedditRunnerUpdateRoute", lago.Route{
		Path:    AppUrl + "workers/{id}/edit/",
		Handler: lago.NewDynamicView("seer_reddit.RedditRunnerUpdateView"),
	})

	_ = lago.RegistryRoute.Register("seer_reddit.RedditRunnerDeleteRoute", lago.Route{
		Path:    AppUrl + "workers/{id}/delete/",
		Handler: lago.NewDynamicView("seer_reddit.RedditRunnerDeleteView"),
	})
}

func init() {
	registerRoutes()
}
