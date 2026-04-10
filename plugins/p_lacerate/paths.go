package p_lacerate

import (
	"github.com/lariv-in/lago/lago"
)

func registerRoutes() {
	_ = lago.RegistryRoute.Register("lacerate.DefaultRoute", lago.Route{
		Path:    AppUrl,
		Handler: lago.NewDynamicView("lacerate.ListView"),
	})

	_ = lago.RegistryRoute.Register("lacerate.CreateRoute", lago.Route{
		Path:    AppUrl + "reddit/sources/create/",
		Handler: lago.NewDynamicView("lacerate.CreateView"),
	})

	_ = lago.RegistryRoute.Register("lacerate.DetailRoute", lago.Route{
		Path:    AppUrl + "reddit/sources/{id}/",
		Handler: lago.NewDynamicView("lacerate.DetailView"),
	})

	_ = lago.RegistryRoute.Register("lacerate.UpdateRoute", lago.Route{
		Path:    AppUrl + "reddit/sources/{id}/edit/",
		Handler: lago.NewDynamicView("lacerate.UpdateView"),
	})

	_ = lago.RegistryRoute.Register("lacerate.DeleteRoute", lago.Route{
		Path:    AppUrl + "reddit/sources/{id}/delete/",
		Handler: lago.NewDynamicView("lacerate.DeleteView"),
	})

	_ = lago.RegistryRoute.Register("lacerate.TwitterDefaultRoute", lago.Route{
		Path:    AppUrl + "twitter/sources/",
		Handler: lago.NewDynamicView("lacerate.TwitterListView"),
	})

	_ = lago.RegistryRoute.Register("lacerate.TwitterCreateRoute", lago.Route{
		Path:    AppUrl + "twitter/sources/create/",
		Handler: lago.NewDynamicView("lacerate.TwitterCreateView"),
	})

	_ = lago.RegistryRoute.Register("lacerate.TwitterDetailRoute", lago.Route{
		Path:    AppUrl + "twitter/sources/{id}/",
		Handler: lago.NewDynamicView("lacerate.TwitterDetailView"),
	})

	_ = lago.RegistryRoute.Register("lacerate.TwitterUpdateRoute", lago.Route{
		Path:    AppUrl + "twitter/sources/{id}/edit/",
		Handler: lago.NewDynamicView("lacerate.TwitterUpdateView"),
	})

	_ = lago.RegistryRoute.Register("lacerate.TwitterDeleteRoute", lago.Route{
		Path:    AppUrl + "twitter/sources/{id}/delete/",
		Handler: lago.NewDynamicView("lacerate.TwitterDeleteView"),
	})

	registerIntelRoutes()
	registerTargetOfInterestRoutes()
	registerLookupRoutes()
}

func registerLookupRoutes() {
	_ = lago.RegistryRoute.Register("lacerate.LookupListRoute", lago.Route{
		Path:    AppUrl + "lookups/",
		Handler: lago.NewDynamicView("lacerate.LookupListView"),
	})

	_ = lago.RegistryRoute.Register("lacerate.LookupCreateRoute", lago.Route{
		Path:    AppUrl + "lookups/create/",
		Handler: lago.NewDynamicView("lacerate.LookupCreateView"),
	})

	_ = lago.RegistryRoute.Register("lacerate.LookupDetailRoute", lago.Route{
		Path:    AppUrl + "lookups/{id}/",
		Handler: lago.NewDynamicView("lacerate.LookupDetailView"),
	})

	_ = lago.RegistryRoute.Register("lacerate.LookupUpdateRoute", lago.Route{
		Path:    AppUrl + "lookups/{id}/edit/",
		Handler: lago.NewDynamicView("lacerate.LookupUpdateView"),
	})

	_ = lago.RegistryRoute.Register("lacerate.LookupDeleteRoute", lago.Route{
		Path:    AppUrl + "lookups/{id}/delete/",
		Handler: lago.NewDynamicView("lacerate.LookupDeleteView"),
	})
}

func registerTargetOfInterestRoutes() {
	_ = lago.RegistryRoute.Register("lacerate.TargetOfInterestListRoute", lago.Route{
		Path:    AppUrl + "targets-of-interest/",
		Handler: lago.NewDynamicView("lacerate.TargetOfInterestListView"),
	})

	_ = lago.RegistryRoute.Register("lacerate.TargetOfInterestCreateRoute", lago.Route{
		Path:    AppUrl + "targets-of-interest/create/",
		Handler: lago.NewDynamicView("lacerate.TargetOfInterestCreateView"),
	})

	_ = lago.RegistryRoute.Register("lacerate.TargetOfInterestDetailRoute", lago.Route{
		Path:    AppUrl + "targets-of-interest/{id}/",
		Handler: lago.NewDynamicView("lacerate.TargetOfInterestDetailView"),
	})

	_ = lago.RegistryRoute.Register("lacerate.TargetOfInterestUpdateRoute", lago.Route{
		Path:    AppUrl + "targets-of-interest/{id}/edit/",
		Handler: lago.NewDynamicView("lacerate.TargetOfInterestUpdateView"),
	})

	_ = lago.RegistryRoute.Register("lacerate.TargetOfInterestDeleteRoute", lago.Route{
		Path:    AppUrl + "targets-of-interest/{id}/delete/",
		Handler: lago.NewDynamicView("lacerate.TargetOfInterestDeleteView"),
	})
}

func registerIntelRoutes() {
	_ = lago.RegistryRoute.Register("lacerate.IntelListRoute", lago.Route{
		Path:    AppUrl + "intel/",
		Handler: lago.NewDynamicView("lacerate.IntelListView"),
	})

	_ = lago.RegistryRoute.Register("lacerate.IntelCreateRoute", lago.Route{
		Path:    AppUrl + "intel/create/",
		Handler: lago.NewDynamicView("lacerate.IntelCreateView"),
	})

	_ = lago.RegistryRoute.Register("lacerate.IntelDetailRoute", lago.Route{
		Path:    AppUrl + "intel/{id}/",
		Handler: lago.NewDynamicView("lacerate.IntelDetailView"),
	})

	_ = lago.RegistryRoute.Register("lacerate.IntelUpdateRoute", lago.Route{
		Path:    AppUrl + "intel/{id}/edit/",
		Handler: lago.NewDynamicView("lacerate.IntelUpdateView"),
	})

	_ = lago.RegistryRoute.Register("lacerate.IntelDeleteRoute", lago.Route{
		Path:    AppUrl + "intel/{id}/delete/",
		Handler: lago.NewDynamicView("lacerate.IntelDeleteView"),
	})

	_ = lago.RegistryRoute.Register("lacerate.SourceSelectRoute", lago.Route{
		Path:    AppUrl + "sources/select/",
		Handler: lago.NewDynamicView("lacerate.SourceSelectView"),
	})
}

func init() {
	registerRoutes()
}
