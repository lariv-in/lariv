package p_nirmancampus_examregistrations

import "github.com/lariv-in/lago/lago"

func registerRoutes() {
	_ = lago.RegistryRoute.Register("examregistrations.DefaultRoute", lago.Route{
		Path:    AppUrl,
		Handler: lago.NewDynamicView("examregistrations.ListView"),
	})

	_ = lago.RegistryRoute.Register("examregistrations.BulkCreateFromAcademicRecordRoute", lago.Route{
		Path:    AppUrl + "bulk-create-academic-record/",
		Handler: lago.NewDynamicView("examregistrations.BulkCreateFromAcademicRecordView"),
	})

	_ = lago.RegistryRoute.Register("examregistrations.DetailRoute", lago.Route{
		Path:    AppUrl + "{id}/",
		Handler: lago.NewDynamicView("examregistrations.DetailView"),
	})

	_ = lago.RegistryRoute.Register("examregistrations.AcademicRecordExamReceiptRoute", lago.Route{
		Path:    AppUrl + "academic-record/{id}/download-receipt/",
		Handler: lago.NewDynamicView("examregistrations.AcademicRecordExamReceiptView"),
	})

	_ = lago.RegistryRoute.Register("examregistrations.UpdateRoute", lago.Route{
		Path:    AppUrl + "{id}/edit/",
		Handler: lago.NewDynamicView("examregistrations.UpdateView"),
	})

	_ = lago.RegistryRoute.Register("examregistrations.DeleteRoute", lago.Route{
		Path:    AppUrl + "{id}/delete/",
		Handler: lago.NewDynamicView("examregistrations.DeleteView"),
	})
}

func init() {
	registerRoutes()
}
