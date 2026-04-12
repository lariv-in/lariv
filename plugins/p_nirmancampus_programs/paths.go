package p_nirmancampus_programs

import (
	"github.com/lariv-in/lago/lago"
)

func registerRoutes() {
	_ = lago.RegistryRoute.Register("programs.DefaultRoute", lago.Route{
		Path:    AppUrl,
		Handler: lago.NewDynamicView("programs.ListView"),
	})

	_ = lago.RegistryRoute.Register("programs.CreateRoute", lago.Route{
		Path:    AppUrl + "create/",
		Handler: lago.NewDynamicView("programs.CreateView"),
	})

	_ = lago.RegistryRoute.Register("programs.DetailRoute", lago.Route{
		Path:    AppUrl + "{id}/",
		Handler: lago.NewDynamicView("programs.DetailView"),
	})

	_ = lago.RegistryRoute.Register("programs.UpdateRoute", lago.Route{
		Path:    AppUrl + "{id}/edit/",
		Handler: lago.NewDynamicView("programs.UpdateView"),
	})

	_ = lago.RegistryRoute.Register("programs.DeleteRoute", lago.Route{
		Path:    AppUrl + "{id}/delete/",
		Handler: lago.NewDynamicView("programs.DeleteView"),
	})

	_ = lago.RegistryRoute.Register("programs.SelectRoute", lago.Route{
		Path:    AppUrl + "select/",
		Handler: lago.NewDynamicView("programs.SelectView"),
	})

	_ = lago.RegistryRoute.Register("programs.ProgramMediaMultiSelectRoute", lago.Route{
		Path:    AppUrl + "program-media/multiselect/",
		Handler: lago.NewDynamicView("programs.ProgramMediaMultiSelectView"),
	})

	_ = lago.RegistryRoute.Register("programs.StructureEditRoute", lago.Route{
		Path:    AppUrl + "{id}/structure/edit/",
		Handler: lago.NewDynamicView("programs.StructureEditView"),
	})

	_ = lago.RegistryRoute.Register("programs.StructureUnitCreateModalRoute", lago.Route{
		Path:    AppUrl + "{id}/structure/units/new/",
		Handler: lago.NewDynamicView("programs.StructureUnitCreateModalView"),
	})

	_ = lago.RegistryRoute.Register("programs.StructureUnitCreateRoute", lago.Route{
		Path:    AppUrl + "{id}/structure/units/",
		Handler: lago.NewDynamicView("programs.StructureUnitCreateView"),
	})

	_ = lago.RegistryRoute.Register("programs.StructureUnitEditModalRoute", lago.Route{
		Path:    AppUrl + "{id}/structure/units/{unitId}/edit/",
		Handler: lago.NewDynamicView("programs.StructureUnitEditModalView"),
	})

	_ = lago.RegistryRoute.Register("programs.StructureUnitUpdateRoute", lago.Route{
		Path:    AppUrl + "{id}/structure/units/{unitId}/",
		Handler: lago.NewDynamicView("programs.StructureUnitUpdateView"),
	})

	_ = lago.RegistryRoute.Register("programs.StructureUnitDeleteRoute", lago.Route{
		Path:    AppUrl + "{id}/structure/units/{unitId}/delete/",
		Handler: lago.NewDynamicView("programs.StructureUnitDeleteView"),
	})
}

func init() {
	registerRoutes()
}
