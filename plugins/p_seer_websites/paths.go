package p_seer_websites

import (
	"github.com/lariv-in/lago/lago"
)

func registerRoutes() {
	_ = lago.RegistryRoute.Register("seer_websites.WebsiteListRoute", lago.Route{
		Path:    AppUrl,
		Handler: lago.NewDynamicView("seer_websites.WebsiteListView"),
	})

	_ = lago.RegistryRoute.Register("seer_websites.WebsiteAddRoute", lago.Route{
		Path:    AppUrl + "add/",
		Handler: lago.NewDynamicView("seer_websites.WebsiteAddView"),
	})

	// Saved scrape rows live under pages/ so /seer-websites/{id}/… does not collide with
	// /seer-websites/workers/{id}/… (Go 1.22+ mux cannot rank those patterns).
	_ = lago.RegistryRoute.Register("seer_websites.WebsiteDetailRoute", lago.Route{
		Path:    AppUrl + "pages/{id}/",
		Handler: lago.NewDynamicView("seer_websites.WebsiteDetailView"),
	})

	_ = lago.RegistryRoute.Register("seer_websites.WebsiteDeleteRoute", lago.Route{
		Path:    AppUrl + "pages/{id}/delete/",
		Handler: lago.NewDynamicView("seer_websites.WebsiteSoftDeleteView"),
	})

	_ = lago.RegistryRoute.Register("seer_websites.WebsiteAddIntelRoute", lago.Route{
		Path:    AppUrl + "pages/{id}/add-intel/",
		Handler: lago.NewDynamicView("seer_websites.WebsiteAddIntelView"),
	})

	_ = lago.RegistryRoute.Register("seer_websites.WebsiteAddAllIntelRoute", lago.Route{
		Path:    AppUrl + "add-all-intel/",
		Handler: lago.NewDynamicView("seer_websites.WebsiteAddAllIntelView"),
	})

	_ = lago.RegistryRoute.Register("seer_websites.WebsiteSourceListRoute", lago.Route{
		Path:    AppUrl + "sources/",
		Handler: lago.NewDynamicView("seer_websites.WebsiteSourceListView"),
	})

	_ = lago.RegistryRoute.Register("seer_websites.WebsiteSourceCreateRoute", lago.Route{
		Path:    AppUrl + "sources/create/",
		Handler: lago.NewDynamicView("seer_websites.WebsiteSourceCreateView"),
	})

	_ = lago.RegistryRoute.Register("seer_websites.WebsiteSourceDetailRoute", lago.Route{
		Path:    AppUrl + "sources/{id}/",
		Handler: lago.NewDynamicView("seer_websites.WebsiteSourceDetailView"),
	})

	_ = lago.RegistryRoute.Register("seer_websites.WebsiteSourceUpdateRoute", lago.Route{
		Path:    AppUrl + "sources/{id}/edit/",
		Handler: lago.NewDynamicView("seer_websites.WebsiteSourceUpdateView"),
	})

	_ = lago.RegistryRoute.Register("seer_websites.WebsiteSourceDeleteRoute", lago.Route{
		Path:    AppUrl + "sources/{id}/delete/",
		Handler: lago.NewDynamicView("seer_websites.WebsiteSourceDeleteView"),
	})

	_ = lago.RegistryRoute.Register("seer_websites.WebsiteSourceFetchRoute", lago.Route{
		Path:    AppUrl + "sources/{source_id}/fetch/",
		Handler: lago.NewDynamicView("seer_websites.WebsiteSourceFetchView"),
	})

	_ = lago.RegistryRoute.Register("seer_websites.WebsiteRunnerListRoute", lago.Route{
		Path:    AppUrl + "workers/",
		Handler: lago.NewDynamicView("seer_websites.WebsiteRunnerListView"),
	})

	_ = lago.RegistryRoute.Register("seer_websites.WebsiteRunnerCreateRoute", lago.Route{
		Path:    AppUrl + "workers/create/",
		Handler: lago.NewDynamicView("seer_websites.WebsiteRunnerCreateView"),
	})

	_ = lago.RegistryRoute.Register("seer_websites.WebsiteRunnerSelectRoute", lago.Route{
		Path:    AppUrl + "workers/select/",
		Handler: lago.NewDynamicView("seer_websites.WebsiteRunnerSelectView"),
	})

	_ = lago.RegistryRoute.Register("seer_websites.WebsiteRunnerWorkerPoolStartRoute", lago.Route{
		Path:    AppUrl + "workers/{id}/worker-pool/start/",
		Handler: lago.NewDynamicView("seer_websites.WebsiteRunnerWorkerPoolStartView"),
	})

	_ = lago.RegistryRoute.Register("seer_websites.WebsiteRunnerWorkerPoolStopRoute", lago.Route{
		Path:    AppUrl + "workers/{id}/worker-pool/stop/",
		Handler: lago.NewDynamicView("seer_websites.WebsiteRunnerWorkerPoolStopView"),
	})

	_ = lago.RegistryRoute.Register("seer_websites.WebsiteRunnerDetailRoute", lago.Route{
		Path:    AppUrl + "workers/{id}/",
		Handler: lago.NewDynamicView("seer_websites.WebsiteRunnerDetailView"),
	})

	_ = lago.RegistryRoute.Register("seer_websites.WebsiteRunnerUpdateRoute", lago.Route{
		Path:    AppUrl + "workers/{id}/edit/",
		Handler: lago.NewDynamicView("seer_websites.WebsiteRunnerUpdateView"),
	})

	_ = lago.RegistryRoute.Register("seer_websites.WebsiteRunnerDeleteRoute", lago.Route{
		Path:    AppUrl + "workers/{id}/delete/",
		Handler: lago.NewDynamicView("seer_websites.WebsiteRunnerDeleteView"),
	})
}

func init() {
	registerRoutes()
}
