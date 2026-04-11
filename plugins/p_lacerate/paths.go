package p_lacerate

import (
	"github.com/lariv-in/lago/lago"
)

func registerRoutes() {
	_ = lago.RegistryRoute.Register("lacerate.DefaultRoute", lago.Route{
		Path:    AppUrl,
		Handler: lago.NewDynamicView("lacerate.SourceListView"),
	})

	_ = lago.RegistryRoute.Register("lacerate.SourceListRoute", lago.Route{
		Path:    AppUrl + "sources/",
		Handler: lago.NewDynamicView("lacerate.SourceListView"),
	})

	_ = lago.RegistryRoute.Register("lacerate.SourceCreateRoute", lago.Route{
		Path:    AppUrl + "sources/create/",
		Handler: lago.NewDynamicView("lacerate.SourceCreateView"),
	})

	_ = lago.RegistryRoute.Register("lacerate.SourceDetailRoute", lago.Route{
		Path:    AppUrl + "sources/{id}/",
		Handler: lago.NewDynamicView("lacerate.SourceDetailView"),
	})

	_ = lago.RegistryRoute.Register("lacerate.SourceUpdateRoute", lago.Route{
		Path:    AppUrl + "sources/{id}/edit/",
		Handler: lago.NewDynamicView("lacerate.SourceUpdateView"),
	})

	_ = lago.RegistryRoute.Register("lacerate.SourceDeleteRoute", lago.Route{
		Path:    AppUrl + "sources/{id}/delete/",
		Handler: lago.NewDynamicView("lacerate.SourceDeleteView"),
	})

	_ = lago.RegistryRoute.Register("lacerate.SourceRestartWorkerRoute", lago.Route{
		Path:    AppUrl + "sources/{id}/restart-worker/",
		Handler: lago.NewDynamicView("lacerate.SourceRestartWorkerView"),
	})

	_ = lago.RegistryRoute.Register("lacerate.SourceStopWorkerRoute", lago.Route{
		Path:    AppUrl + "sources/{id}/stop-worker/",
		Handler: lago.NewDynamicView("lacerate.SourceStopWorkerView"),
	})

	registerIntelRoutes()
	registerReportRoutes()
	registerTargetOfInterestRoutes()
	registerLookupRoutes()
}

func registerTargetOfInterestRoutes() {
	_ = lago.RegistryRoute.Register("lacerate.TargetOfInterestListRoute", lago.Route{
		Path:    AppUrl + "toi/",
		Handler: lago.NewDynamicView("lacerate.TargetOfInterestListView"),
	})

	_ = lago.RegistryRoute.Register("lacerate.TargetOfInterestCreateRoute", lago.Route{
		Path:    AppUrl + "toi/create/",
		Handler: lago.NewDynamicView("lacerate.TargetOfInterestCreateView"),
	})

	_ = lago.RegistryRoute.Register("lacerate.TargetOfInterestDetailRoute", lago.Route{
		Path:    AppUrl + "toi/{id}/",
		Handler: lago.NewDynamicView("lacerate.TargetOfInterestDetailView"),
	})

	_ = lago.RegistryRoute.Register("lacerate.TargetOfInterestUpdateRoute", lago.Route{
		Path:    AppUrl + "toi/{id}/edit/",
		Handler: lago.NewDynamicView("lacerate.TargetOfInterestUpdateView"),
	})

	_ = lago.RegistryRoute.Register("lacerate.TargetOfInterestDeleteRoute", lago.Route{
		Path:    AppUrl + "toi/{id}/delete/",
		Handler: lago.NewDynamicView("lacerate.TargetOfInterestDeleteView"),
	})
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

	_ = lago.RegistryRoute.Register("lacerate.LookupRestartWorkerRoute", lago.Route{
		Path:    AppUrl + "lookups/{id}/restart-worker/",
		Handler: lago.NewDynamicView("lacerate.LookupRestartWorkerView"),
	})

	_ = lago.RegistryRoute.Register("lacerate.LookupStopWorkerRoute", lago.Route{
		Path:    AppUrl + "lookups/{id}/stop-worker/",
		Handler: lago.NewDynamicView("lacerate.LookupStopWorkerView"),
	})
}

func registerReportRoutes() {
	_ = lago.RegistryRoute.Register("lacerate.ReportListRoute", lago.Route{
		Path:    AppUrl + "reports/",
		Handler: lago.NewDynamicView("lacerate.ReportListView"),
	})

	_ = lago.RegistryRoute.Register("lacerate.ReportCreateRoute", lago.Route{
		Path:    AppUrl + "reports/create/",
		Handler: lago.NewDynamicView("lacerate.ReportCreateView"),
	})

	_ = lago.RegistryRoute.Register("lacerate.ReportDetailRoute", lago.Route{
		Path:    AppUrl + "reports/{id}/",
		Handler: lago.NewDynamicView("lacerate.ReportDetailView"),
	})

	_ = lago.RegistryRoute.Register("lacerate.ReportUpdateRoute", lago.Route{
		Path:    AppUrl + "reports/{id}/edit/",
		Handler: lago.NewDynamicView("lacerate.ReportUpdateView"),
	})

	_ = lago.RegistryRoute.Register("lacerate.ReportDeleteRoute", lago.Route{
		Path:    AppUrl + "reports/{id}/delete/",
		Handler: lago.NewDynamicView("lacerate.ReportDeleteView"),
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
