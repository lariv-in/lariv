package p_seer_aisstream

import (
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
)

func init() {
	lago.RegistryView.Register("seer_aisstream.MapView",
		lago.GetPageView("seer_aisstream.MapPage").
			WithLayer("users.auth", p_users.AuthenticationLayer{}))
}
