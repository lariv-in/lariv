package p_nirmancampus_sessions

import (
	"github.com/lariv-in/lago/lago"
)

func registerRoutes() {
	_ = lago.RegistryRoute.Register("sessions.DefaultRoute", lago.Route{
		Path:    AppUrl,
		Handler: lago.NewDynamicView("sessions.ListView"),
	})

	_ = lago.RegistryRoute.Register("sessions.CreateRoute", lago.Route{
		Path:    AppUrl + "admissionsessions/create/",
		Handler: lago.NewDynamicView("sessions.CreateView"),
	})

	_ = lago.RegistryRoute.Register("sessions.DetailRoute", lago.Route{
		Path:    AppUrl + "admissionsessions/{id}/",
		Handler: lago.NewDynamicView("sessions.DetailView"),
	})

	_ = lago.RegistryRoute.Register("sessions.UpdateRoute", lago.Route{
		Path:    AppUrl + "admissionsessions/{id}/edit/",
		Handler: lago.NewDynamicView("sessions.UpdateView"),
	})

	_ = lago.RegistryRoute.Register("sessions.DeleteRoute", lago.Route{
		Path:    AppUrl + "admissionsessions/{id}/delete/",
		Handler: lago.NewDynamicView("sessions.DeleteView"),
	})

	_ = lago.RegistryRoute.Register("sessions.SelectRoute", lago.Route{
		Path:    AppUrl + "admissionsessions/select/",
		Handler: lago.NewDynamicView("sessions.SelectView"),
	})

	_ = lago.RegistryRoute.Register("sessions.ExamCreateRoute", lago.Route{
		Path:    AppUrl + "examsessions/new/",
		Handler: lago.NewDynamicView("sessions.ExamCreateView"),
	})

	_ = lago.RegistryRoute.Register("sessions.ExamDetailRoute", lago.Route{
		Path:    AppUrl + "examsessions/{id}/",
		Handler: lago.NewDynamicView("sessions.ExamDetailView"),
	})

	_ = lago.RegistryRoute.Register("sessions.ExamUpdateRoute", lago.Route{
		Path:    AppUrl + "examsessions/{id}/edit/",
		Handler: lago.NewDynamicView("sessions.ExamUpdateView"),
	})

	_ = lago.RegistryRoute.Register("sessions.ExamDeleteRoute", lago.Route{
		Path:    AppUrl + "examsessions/{id}/delete/",
		Handler: lago.NewDynamicView("sessions.ExamDeleteView"),
	})
}

func init() {
	registerRoutes()
}
