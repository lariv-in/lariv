package p_export

import "github.com/lariv-in/lago/lago"

func init() {
	_ = lago.RegistryRoute.Register("export.PageRoute", lago.Route{
		Path:    AppUrl,
		Handler: lago.NewDynamicView("export.PageView"),
	})
	_ = lago.RegistryRoute.Register("export.DownloadRoute", lago.Route{
		Path:    AppUrl + "download/",
		Handler: lago.NewDynamicView("export.DownloadView"),
	})
}
