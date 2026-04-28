package p_nirmancampus_academicrecords

import "github.com/lariv-in/lago/lago"

func registerRoutes() {
	_ = lago.RegistryRoute.Register("academicrecords.DefaultRoute", lago.Route{
		Path:    AppUrl,
		Handler: lago.NewDynamicView("academicrecords.ListView"),
	})

	_ = lago.RegistryRoute.Register("academicrecords.CreateRoute", lago.Route{
		Path:    AppUrl + "create/",
		Handler: lago.NewDynamicView("academicrecords.CreateView"),
	})

	_ = lago.RegistryRoute.Register("academicrecords.DetailRoute", lago.Route{
		Path:    AppUrl + "{id}/",
		Handler: lago.NewDynamicView("academicrecords.DetailView"),
	})

	_ = lago.RegistryRoute.Register("academicrecords.DownloadPdfRoute", lago.Route{
		Path:    AppUrl + "{id}/download-pdf/",
		Handler: lago.NewDynamicView("academicrecords.DownloadPdfView"),
	})

	_ = lago.RegistryRoute.Register("academicrecords.UpdateRoute", lago.Route{
		Path:    AppUrl + "{id}/edit/",
		Handler: lago.NewDynamicView("academicrecords.UpdateView"),
	})

	_ = lago.RegistryRoute.Register("academicrecords.DeleteRoute", lago.Route{
		Path:    AppUrl + "{id}/delete/",
		Handler: lago.NewDynamicView("academicrecords.DeleteView"),
	})

	_ = lago.RegistryRoute.Register("academicrecords.SelectRoute", lago.Route{
		Path:    AppUrl + "select/",
		Handler: lago.NewDynamicView("academicrecords.SelectView"),
	})

	_ = lago.RegistryRoute.Register("academicrecords.ProgramStructureUnitSelectRoute", lago.Route{
		Path:    AppUrl + "program-structure-units/select/",
		Handler: lago.NewDynamicView("academicrecords.ProgramStructureUnitSelectView"),
	})
}

func init() {
	registerRoutes()
}
