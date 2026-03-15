package p_totschool_appointments

import (
	"github.com/lariv-in/lago"
)

func init() {
	_ = lago.RegistryRoute.Register("appointments.ListRoute", lago.Route{
		Path:    AppUrl,
		Handler: lago.NewDynamicView("appointments.ListView"),
	})
	_ = lago.RegistryRoute.Register("appointments.CreateRoute", lago.Route{
		Path:    AppUrl + "create/",
		Handler: lago.NewDynamicView("appointments.CreateView"),
	})
	_ = lago.RegistryRoute.Register("appointments.DetailRoute", lago.Route{
		Path:    AppUrl + "{id}/",
		Handler: lago.NewDynamicView("appointments.DetailView"),
	})
	_ = lago.RegistryRoute.Register("appointments.UpdateRoute", lago.Route{
		Path:    AppUrl + "{id}/edit/",
		Handler: lago.NewDynamicView("appointments.UpdateView"),
	})
	_ = lago.RegistryRoute.Register("appointments.DeleteRoute", lago.Route{
		Path:    AppUrl + "{id}/delete/",
		Handler: lago.NewDynamicView("appointments.DeleteView"),
	})
	_ = lago.RegistryRoute.Register("appointments.GenerateRoute", lago.Route{
		Path:    AppUrl + "{id}/generate/",
		Handler: lago.NewDynamicView("appointments.GenerateView"),
	})
	_ = lago.RegistryRoute.Register("appointments.CancelRoute", lago.Route{
		Path:    AppUrl + "{id}/cancel/",
		Handler: lago.NewDynamicView("appointments.CancelView"),
	})
	_ = lago.RegistryRoute.Register("appointments.AiEditFormRoute", lago.Route{
		Path:    AppUrl + "{id}/ai-edit/form/",
		Handler: lago.NewDynamicView("appointments.AiEditFormView"),
	})
	_ = lago.RegistryRoute.Register("appointments.AiEditRoute", lago.Route{
		Path:    AppUrl + "{id}/ai-edit/",
		Handler: lago.NewDynamicView("appointments.AiEditView"),
	})
	_ = lago.RegistryRoute.Register("appointments.SelectRoute", lago.Route{
		Path:    AppUrl + "select/",
		Handler: lago.NewDynamicView("appointments.SelectView"),
	})
	_ = lago.RegistryRoute.Register("appointments.CardTimelineRoute", lago.Route{
		Path:    AppUrl + "cards/",
		Handler: lago.NewDynamicView("appointments.CardTimelineView"),
	})
}
