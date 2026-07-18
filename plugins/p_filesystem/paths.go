package p_filesystem

import (
	"net/http"

	"github.com/lariv-in/lariv"
	"github.com/lariv-in/lariv/registry"
)

func pluginRoutes() lariv.PluginFeatures[lariv.Route] {
	return lariv.PluginFeatures[lariv.Route]{
		Entries: []registry.Pair[string, lariv.Route]{
			{Key: "filesystem.ListRoute", Value: lariv.Route{
				Path:    AppUrl,
				Handler: lariv.NewDynamicView("filesystem.ListView"),
			}},
			{Key: "filesystem.BrowseRoute", Value: lariv.Route{
				Path:    AppUrl + "browse/in/{parent_id}/",
				Handler: lariv.NewDynamicView("filesystem.BrowseView"),
			}},
			{Key: "filesystem.SelectRoute", Value: lariv.Route{
				Path:    AppUrl + "select/",
				Handler: lariv.NewDynamicView("filesystem.SelectView"),
			}},
			{Key: "filesystem.SelectChildRoute", Value: lariv.Route{
				Path:    AppUrl + "select/in/{parent_id}/",
				Handler: lariv.NewDynamicView("filesystem.SelectChildView"),
			}},
			{Key: "filesystem.FileSelectRoute", Value: lariv.Route{
				Path:    AppUrl + "select-file/",
				Handler: lariv.NewDynamicView("filesystem.FileSelectView"),
			}},
			{Key: "filesystem.FileSelectChildRoute", Value: lariv.Route{
				Path:    AppUrl + "select-file/in/{parent_id}/",
				Handler: lariv.NewDynamicView("filesystem.FileSelectChildView"),
			}},
			{Key: "filesystem.MultiSelectRoute", Value: lariv.Route{
				Path:    AppUrl + "multi-select/",
				Handler: lariv.NewDynamicView("filesystem.MultiSelectView"),
			}},
			{Key: "filesystem.MultiSelectChildRoute", Value: lariv.Route{
				Path:    AppUrl + "multi-select/in/{parent_id}/",
				Handler: lariv.NewDynamicView("filesystem.MultiSelectChildView"),
			}},
			{Key: "filesystem.MoveSelectRoute", Value: lariv.Route{
				Path:    AppUrl + "move-select/",
				Handler: lariv.NewDynamicView("filesystem.MoveSelectView"),
			}},
			{Key: "filesystem.MoveSelectChildRoute", Value: lariv.Route{
				Path:    AppUrl + "move-select/in/{parent_id}/",
				Handler: lariv.NewDynamicView("filesystem.MoveSelectChildView"),
			}},
			{Key: "filesystem.CreateRoute", Value: lariv.Route{
				Path:    AppUrl + "create/",
				Handler: lariv.NewDynamicView("filesystem.CreateView"),
			}},
			{Key: "filesystem.CreateChildRoute", Value: lariv.Route{
				Path:    AppUrl + "create/in/{parent_id}/",
				Handler: lariv.NewDynamicView("filesystem.CreateChildView"),
			}},
			{Key: "filesystem.MultiUploadRoute", Value: lariv.Route{
				Path:    AppUrl + "upload/",
				Handler: lariv.NewDynamicView("filesystem.MultiUploadView"),
			}},
			{Key: "filesystem.MultiUploadChildRoute", Value: lariv.Route{
				Path:    AppUrl + "upload/in/{parent_id}/",
				Handler: lariv.NewDynamicView("filesystem.MultiUploadChildView"),
			}},
			{Key: "filesystem.ZipUploadRoute", Value: lariv.Route{
				Path:    AppUrl + "zip-upload/",
				Handler: lariv.NewDynamicView("filesystem.ZipUploadView"),
			}},
			{Key: "filesystem.ZipUploadChildRoute", Value: lariv.Route{
				Path:    AppUrl + "zip-upload/in/{parent_id}/",
				Handler: lariv.NewDynamicView("filesystem.ZipUploadChildView"),
			}},
			{Key: "filesystem.DetailRoute", Value: lariv.Route{
				Path:    AppUrl + "{id}/",
				Handler: lariv.NewDynamicView("filesystem.DetailView"),
			}},
			{Key: "filesystem.DownloadRoute", Value: lariv.Route{
				Path:    AppUrl + "{id}/download/",
				Handler: lariv.NewDynamicView("filesystem.DownloadView"),
			}},
			{Key: "filesystem.DownloadRootRoute", Value: lariv.Route{
				Path:    AppUrl + "download/",
				Handler: lariv.NewDynamicView("filesystem.DownloadRootView"),
			}},
			{Key: "filesystem.UpdateRoute", Value: lariv.Route{
				Path:    AppUrl + "{id}/edit/",
				Handler: lariv.NewDynamicView("filesystem.UpdateView"),
			}},
			{Key: "filesystem.DeleteRoute", Value: lariv.Route{
				Path:    AppUrl + "{id}/delete/",
				Handler: lariv.NewDynamicView("filesystem.DeleteView"),
			}},
			{Key: "filesystem.MoveRoute", Value: lariv.Route{
				Path:    AppUrl + "{id}/move/",
				Handler: lariv.NewDynamicView("filesystem.MoveView"),
			}},
			{Key: "filesystem.ChatUploadRoute", Value: lariv.Route{
				Path:    AppUrl + "chat-upload/",
				Handler: http.HandlerFunc(chatUploadHandler),
			}},
		},
	}
}
