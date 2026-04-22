package p_google_genai

import "github.com/lariv-in/lago/lago"

func init() {
	_ = lago.RegistryRoute.Register("googlegenai.PageRoute", lago.Route{
		Path:    AppUrl,
		Handler: lago.NewDynamicView("googlegenai.PageView"),
	})
}
