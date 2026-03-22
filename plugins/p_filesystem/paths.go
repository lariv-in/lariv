package p_filesystem

import "github.com/lariv-in/lago/lago"

func init() {
	_ = lago.RegistryRoute.Register("filesystem.ListRoute", lago.Route{
		Path:    AppUrl,
		Handler: lago.NewDynamicView("filesystem.ListView"),
	})
	_ = lago.RegistryRoute.Register("filesystem.BrowseRoute", lago.Route{
		Path:    AppUrl + "browse/in/{parent_id}/",
		Handler: lago.NewDynamicView("filesystem.BrowseView"),
	})
	_ = lago.RegistryRoute.Register("filesystem.SelectRoute", lago.Route{
		Path:    AppUrl + "select/",
		Handler: lago.NewDynamicView("filesystem.SelectView"),
	})
	_ = lago.RegistryRoute.Register("filesystem.SelectChildRoute", lago.Route{
		Path:    AppUrl + "select/in/{parent_id}/",
		Handler: lago.NewDynamicView("filesystem.SelectChildView"),
	})
	_ = lago.RegistryRoute.Register("filesystem.MultiSelectRoute", lago.Route{
		Path:    AppUrl + "multi-select/",
		Handler: lago.NewDynamicView("filesystem.MultiSelectView"),
	})
	_ = lago.RegistryRoute.Register("filesystem.MultiSelectChildRoute", lago.Route{
		Path:    AppUrl + "multi-select/in/{parent_id}/",
		Handler: lago.NewDynamicView("filesystem.MultiSelectChildView"),
	})
	_ = lago.RegistryRoute.Register("filesystem.MoveSelectRoute", lago.Route{
		Path:    AppUrl + "move-select/",
		Handler: lago.NewDynamicView("filesystem.MoveSelectView"),
	})
	_ = lago.RegistryRoute.Register("filesystem.MoveSelectChildRoute", lago.Route{
		Path:    AppUrl + "move-select/in/{parent_id}/",
		Handler: lago.NewDynamicView("filesystem.MoveSelectChildView"),
	})
	_ = lago.RegistryRoute.Register("filesystem.CreateRoute", lago.Route{
		Path:    AppUrl + "create/",
		Handler: lago.NewDynamicView("filesystem.CreateView"),
	})
	_ = lago.RegistryRoute.Register("filesystem.CreateChildRoute", lago.Route{
		Path:    AppUrl + "create/in/{parent_id}/",
		Handler: lago.NewDynamicView("filesystem.CreateChildView"),
	})
	_ = lago.RegistryRoute.Register("filesystem.MultiUploadRoute", lago.Route{
		Path:    AppUrl + "upload/",
		Handler: lago.NewDynamicView("filesystem.MultiUploadView"),
	})
	_ = lago.RegistryRoute.Register("filesystem.MultiUploadChildRoute", lago.Route{
		Path:    AppUrl + "upload/in/{parent_id}/",
		Handler: lago.NewDynamicView("filesystem.MultiUploadChildView"),
	})
	_ = lago.RegistryRoute.Register("filesystem.DetailRoute", lago.Route{
		Path:    AppUrl + "{id}/",
		Handler: lago.NewDynamicView("filesystem.DetailView"),
	})
	_ = lago.RegistryRoute.Register("filesystem.DownloadRoute", lago.Route{
		Path:    AppUrl + "{id}/download/",
		Handler: lago.NewDynamicView("filesystem.DownloadView"),
	})
	_ = lago.RegistryRoute.Register("filesystem.UpdateRoute", lago.Route{
		Path:    AppUrl + "{id}/edit/",
		Handler: lago.NewDynamicView("filesystem.UpdateView"),
	})
	_ = lago.RegistryRoute.Register("filesystem.DeleteRoute", lago.Route{
		Path:    AppUrl + "{id}/delete/",
		Handler: lago.NewDynamicView("filesystem.DeleteView"),
	})
	_ = lago.RegistryRoute.Register("filesystem.MoveRoute", lago.Route{
		Path:    AppUrl + "{id}/move/",
		Handler: lago.NewDynamicView("filesystem.MoveView"),
	})
}
