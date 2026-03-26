package p_nirmancampus_assignmentresults

import (
	"github.com/lariv-in/lago/lago"
)

// ResultsAppUrl is the URL prefix for this plugin. It is not nested under /assignments/… so it
// cannot conflict with /assignments/{id}/, /assignments/{id}/edit/, etc. (see net/http ServeMux).
const ResultsAppUrl = "/assignmentresults/"

func registerRoutes() {
	_ = lago.RegistryRoute.Register("assignmentresults.DefaultRoute", lago.Route{
		Path:    ResultsAppUrl,
		Handler: lago.NewDynamicView("assignmentresults.ListView"),
	})

	_ = lago.RegistryRoute.Register("assignmentresults.CreateRoute", lago.Route{
		Path:    ResultsAppUrl + "create/",
		Handler: lago.NewDynamicView("assignmentresults.CreateView"),
	})

	_ = lago.RegistryRoute.Register("assignmentresults.DetailRoute", lago.Route{
		Path:    ResultsAppUrl + "{id}/",
		Handler: lago.NewDynamicView("assignmentresults.DetailView"),
	})

	_ = lago.RegistryRoute.Register("assignmentresults.UpdateRoute", lago.Route{
		Path:    ResultsAppUrl + "{id}/edit/",
		Handler: lago.NewDynamicView("assignmentresults.UpdateView"),
	})

	_ = lago.RegistryRoute.Register("assignmentresults.DeleteRoute", lago.Route{
		Path:    ResultsAppUrl + "{id}/delete/",
		Handler: lago.NewDynamicView("assignmentresults.DeleteView"),
	})

	_ = lago.RegistryRoute.Register("assignmentresults.SelectRoute", lago.Route{
		Path:    ResultsAppUrl + "select/",
		Handler: lago.NewDynamicView("assignmentresults.SelectView"),
	})
}

func init() {
	registerRoutes()
}
