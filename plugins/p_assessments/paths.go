package p_assessments

import "github.com/lariv-in/lago/lago"

func registerRoutes() {
	_ = lago.RegistryRoute.Register("assessments.DefaultRoute", lago.Route{Path: AppURL, Handler: lago.NewDynamicView("assessments.ListView")})
	_ = lago.RegistryRoute.Register("assessments.CreateRoute", lago.Route{Path: AppURL + "create/", Handler: lago.NewDynamicView("assessments.CreateView")})
	_ = lago.RegistryRoute.Register("assessments.DetailRoute", lago.Route{Path: AppURL + "{id}/", Handler: lago.NewDynamicView("assessments.DetailView")})
	_ = lago.RegistryRoute.Register("assessments.UpdateRoute", lago.Route{Path: AppURL + "{id}/edit/", Handler: lago.NewDynamicView("assessments.UpdateView")})
	_ = lago.RegistryRoute.Register("assessments.DeleteRoute", lago.Route{Path: AppURL + "{id}/delete/", Handler: lago.NewDynamicView("assessments.DeleteView")})
	_ = lago.RegistryRoute.Register("assessments.ExamDefaultRoute", lago.Route{Path: ExamsURL, Handler: lago.NewDynamicView("assessments.ExamListView")})
	_ = lago.RegistryRoute.Register("assessments.ExamCreateRoute", lago.Route{Path: ExamsURL + "create/", Handler: lago.NewDynamicView("assessments.ExamCreateView")})
	_ = lago.RegistryRoute.Register("assessments.ExamDetailRoute", lago.Route{Path: ExamsURL + "{id}/", Handler: lago.NewDynamicView("assessments.ExamDetailView")})
	_ = lago.RegistryRoute.Register("assessments.ExamUpdateRoute", lago.Route{Path: ExamsURL + "{id}/edit/", Handler: lago.NewDynamicView("assessments.ExamUpdateView")})
	_ = lago.RegistryRoute.Register("assessments.ExamDeleteRoute", lago.Route{Path: ExamsURL + "{id}/delete/", Handler: lago.NewDynamicView("assessments.ExamDeleteView")})
}

func init() { registerRoutes() }
