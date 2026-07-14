package p_filesystem

import (
	"net/http"

	"github.com/lariv-in/lago"
	"github.com/lariv-in/lago/registry"
)

func pluginRoutes() lago.PluginFeatures[lago.Route] {
	return lago.PluginFeatures[lago.Route]{
		Entries: []registry.Pair[string, lago.Route]{
			{Key: "filesystem.ListRoute", Value: lago.Route{
				Path:    AppUrl,
				Handler: lago.NewDynamicView("filesystem.ListView"),
			}},
			{Key: "filesystem.BrowseRoute", Value: lago.Route{
				Path:    AppUrl + "browse/in/{parent_id}/",
				Handler: lago.NewDynamicView("filesystem.BrowseView"),
			}},
			{Key: "filesystem.SelectRoute", Value: lago.Route{
				Path:    AppUrl + "select/",
				Handler: lago.NewDynamicView("filesystem.SelectView"),
			}},
			{Key: "filesystem.SelectChildRoute", Value: lago.Route{
				Path:    AppUrl + "select/in/{parent_id}/",
				Handler: lago.NewDynamicView("filesystem.SelectChildView"),
			}},
			{Key: "filesystem.MultiSelectRoute", Value: lago.Route{
				Path:    AppUrl + "multi-select/",
				Handler: lago.NewDynamicView("filesystem.MultiSelectView"),
			}},
			{Key: "filesystem.MultiSelectChildRoute", Value: lago.Route{
				Path:    AppUrl + "multi-select/in/{parent_id}/",
				Handler: lago.NewDynamicView("filesystem.MultiSelectChildView"),
			}},
			{Key: "filesystem.MoveSelectRoute", Value: lago.Route{
				Path:    AppUrl + "move-select/",
				Handler: lago.NewDynamicView("filesystem.MoveSelectView"),
			}},
			{Key: "filesystem.MoveSelectChildRoute", Value: lago.Route{
				Path:    AppUrl + "move-select/in/{parent_id}/",
				Handler: lago.NewDynamicView("filesystem.MoveSelectChildView"),
			}},
			{Key: "filesystem.CreateRoute", Value: lago.Route{
				Path:    AppUrl + "create/",
				Handler: lago.NewDynamicView("filesystem.CreateView"),
			}},
			{Key: "filesystem.CreateChildRoute", Value: lago.Route{
				Path:    AppUrl + "create/in/{parent_id}/",
				Handler: lago.NewDynamicView("filesystem.CreateChildView"),
			}},
			{Key: "filesystem.MultiUploadRoute", Value: lago.Route{
				Path:    AppUrl + "upload/",
				Handler: lago.NewDynamicView("filesystem.MultiUploadView"),
			}},
			{Key: "filesystem.MultiUploadChildRoute", Value: lago.Route{
				Path:    AppUrl + "upload/in/{parent_id}/",
				Handler: lago.NewDynamicView("filesystem.MultiUploadChildView"),
			}},
			{Key: "filesystem.DetailRoute", Value: lago.Route{
				Path:    AppUrl + "{id}/",
				Handler: lago.NewDynamicView("filesystem.DetailView"),
			}},
			{Key: "filesystem.DownloadRoute", Value: lago.Route{
				Path:    AppUrl + "{id}/download/",
				Handler: lago.NewDynamicView("filesystem.DownloadView"),
			}},
			{Key: "filesystem.UpdateRoute", Value: lago.Route{
				Path:    AppUrl + "{id}/edit/",
				Handler: lago.NewDynamicView("filesystem.UpdateView"),
			}},
			{Key: "filesystem.DeleteRoute", Value: lago.Route{
				Path:    AppUrl + "{id}/delete/",
				Handler: lago.NewDynamicView("filesystem.DeleteView"),
			}},
			{Key: "filesystem.MoveRoute", Value: lago.Route{
				Path:    AppUrl + "{id}/move/",
				Handler: lago.NewDynamicView("filesystem.MoveView"),
			}},
			{Key: "filesystem.ChatUploadRoute", Value: lago.Route{
				Path:    AppUrl + "chat-upload/",
				Handler: http.HandlerFunc(chatUploadHandler),
			}},
		},
	}
}
