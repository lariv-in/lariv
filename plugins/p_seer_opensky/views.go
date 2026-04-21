package p_seer_opensky

import (
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
)

func init() {
	lago.RegistryView.Register("seer_opensky.MapView",
		lago.GetPageView("seer_opensky.MapPage").
			WithLayer("users.auth", p_users.AuthenticationLayer{}))
}
