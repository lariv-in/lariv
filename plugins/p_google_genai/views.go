package p_google_genai

import (
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
)

func init() {
	lago.RegistryView.Register("googlegenai.PageView",
		lago.GetPageView("googlegenai.Page").
			WithLayer("users.auth", p_users.AuthenticationLayer{}))
}
