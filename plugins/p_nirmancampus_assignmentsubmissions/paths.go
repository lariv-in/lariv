package p_nirmancampus_assignmentsubmissions

import "github.com/lariv-in/lago/lago"

func registerRoutes() {
	_ = lago.RegistryRoute.Register("assignmentsubmissions.DefaultRoute", lago.Route{
		Path:    AppUrl,
		Handler: lago.NewDynamicView("assignmentsubmissions.ListView"),
	})

	_ = lago.RegistryRoute.Register("assignmentsubmissions.BulkCreateFromAcademicRecordRoute", lago.Route{
		Path:    AppUrl + "bulk-create-academic-record/",
		Handler: lago.NewDynamicView("assignmentsubmissions.BulkCreateFromAcademicRecordView"),
	})

	_ = lago.RegistryRoute.Register("assignmentsubmissions.DetailRoute", lago.Route{
		Path:    AppUrl + "{id}/",
		Handler: lago.NewDynamicView("assignmentsubmissions.DetailView"),
	})

	_ = lago.RegistryRoute.Register("assignmentsubmissions.UpdateRoute", lago.Route{
		Path:    AppUrl + "{id}/edit/",
		Handler: lago.NewDynamicView("assignmentsubmissions.UpdateView"),
	})

	_ = lago.RegistryRoute.Register("assignmentsubmissions.DeleteRoute", lago.Route{
		Path:    AppUrl + "{id}/delete/",
		Handler: lago.NewDynamicView("assignmentsubmissions.DeleteView"),
	})
}

func init() {
	registerRoutes()
}
