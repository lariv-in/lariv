package p_totschool_proposals

import (
	"github.com/lariv-in/lago/lago"
)

func init() {
	_ = lago.RegistryRoute.Register("proposals.ListRoute", lago.Route{
		Path:    AppUrl,
		Handler: lago.NewDynamicView("proposals.ListView"),
	})
	_ = lago.RegistryRoute.Register("proposals.CreateRoute", lago.Route{
		Path:    AppUrl + "create/",
		Handler: lago.NewDynamicView("proposals.CreateView"),
	})
	_ = lago.RegistryRoute.Register("proposals.DetailRoute", lago.Route{
		Path:    AppUrl + "{id}/",
		Handler: lago.NewDynamicView("proposals.DetailView"),
	})
	_ = lago.RegistryRoute.Register("proposals.UpdateRoute", lago.Route{
		Path:    AppUrl + "{id}/edit/",
		Handler: lago.NewDynamicView("proposals.UpdateView"),
	})
	_ = lago.RegistryRoute.Register("proposals.DeleteRoute", lago.Route{
		Path:    AppUrl + "{id}/delete/",
		Handler: lago.NewDynamicView("proposals.DeleteView"),
	})
	_ = lago.RegistryRoute.Register("proposals.GenerateRoute", lago.Route{
		Path:    AppUrl + "{id}/generate/",
		Handler: lago.NewDynamicView("proposals.GenerateView"),
	})
	_ = lago.RegistryRoute.Register("proposals.CancelRoute", lago.Route{
		Path:    AppUrl + "{id}/cancel/",
		Handler: lago.NewDynamicView("proposals.CancelView"),
	})
	_ = lago.RegistryRoute.Register("proposals.AiEditFormRoute", lago.Route{
		Path:    AppUrl + "{id}/ai-edit/form/",
		Handler: lago.NewDynamicView("proposals.AiEditFormView"),
	})
	_ = lago.RegistryRoute.Register("proposals.AiEditRoute", lago.Route{
		Path:    AppUrl + "{id}/ai-edit/",
		Handler: lago.NewDynamicView("proposals.AiEditView"),
	})
	_ = lago.RegistryRoute.Register("proposals.ExportPdfRoute", lago.Route{
		Path:    AppUrl + "{id}/export-pdf/",
		Handler: lago.NewDynamicView("proposals.ExportPdfView"),
	})
	_ = lago.RegistryRoute.Register("proposals.ExportDocxRoute", lago.Route{
		Path:    AppUrl + "{id}/export-docx/",
		Handler: lago.NewDynamicView("proposals.ExportDocxView"),
	})
}
