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

	_ = lago.RegistryRoute.Register("seer_websites.WebsiteDetailRoute", lago.Route{
		Path:    AppUrl + "{id}/",
		Handler: lago.NewDynamicView("seer_websites.WebsiteDetailView"),
	})

	_ = lago.RegistryRoute.Register("seer_websites.WebsiteDeleteRoute", lago.Route{
		Path:    AppUrl + "{id}/delete/",
		Handler: lago.NewDynamicView("seer_websites.WebsiteSoftDeleteView"),
	})

	_ = lago.RegistryRoute.Register("seer_websites.WebsiteAddIntelRoute", lago.Route{
		Path:    AppUrl + "{id}/add-intel/",
		Handler: lago.NewDynamicView("seer_websites.WebsiteAddIntelView"),
	})

	_ = lago.RegistryRoute.Register("seer_websites.WebsiteAddAllIntelRoute", lago.Route{
		Path:    AppUrl + "add-all-intel/",
		Handler: lago.NewDynamicView("seer_websites.WebsiteAddAllIntelView"),
	})
}

func init() {
	registerRoutes()
}
