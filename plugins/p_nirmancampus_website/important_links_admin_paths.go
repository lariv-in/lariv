package p_nirmancampus_website

import "github.com/lariv-in/lago/lago"

const importantLinksAdminUrl = AppUrl + "important-links/"

func init() {
	// --- Important links routes ---
	_ = lago.RegistryRoute.Register("nirmancampus_website.ImportantLinksDefaultRoute", lago.Route{
		Path:    importantLinksAdminUrl,
		Handler: lago.NewDynamicView("nirmancampus_website.ImportantLinksListView"),
	})

	_ = lago.RegistryRoute.Register("nirmancampus_website.ImportantLinksCreateRoute", lago.Route{
		Path:    importantLinksAdminUrl + "create/",
		Handler: lago.NewDynamicView("nirmancampus_website.ImportantLinksCreateView"),
	})

	_ = lago.RegistryRoute.Register("nirmancampus_website.ImportantLinksDetailRoute", lago.Route{
		Path:    importantLinksAdminUrl + "{id}/",
		Handler: lago.NewDynamicView("nirmancampus_website.ImportantLinksDetailView"),
	})

	_ = lago.RegistryRoute.Register("nirmancampus_website.ImportantLinksUpdateRoute", lago.Route{
		Path:    importantLinksAdminUrl + "{id}/edit/",
		Handler: lago.NewDynamicView("nirmancampus_website.ImportantLinksUpdateView"),
	})

	_ = lago.RegistryRoute.Register("nirmancampus_website.ImportantLinksDeleteRoute", lago.Route{
		Path:    importantLinksAdminUrl + "{id}/delete/",
		Handler: lago.NewDynamicView("nirmancampus_website.ImportantLinksDeleteView"),
	})
}

