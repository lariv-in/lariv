package p_otp

import (
	"log"
	"net/url"

	"github.com/lariv-in/lago"
)

const AppUrl = "/otp/preferences/"

func init() {
	u, err := url.Parse(AppUrl)
	if err != nil {
		log.Panic(err)
	}

	err = lago.RegistryPlugins.Register("p_otp", lago.Plugin{
		Type:        lago.PluginTypeApp,
		Icon:        "key", // Using a relevant icon
		Url:         u,
		VerboseName: "OTP Preferences",
		RenderKeys:  []string{"superuser", "totschool_admin"},
	})
	if err != nil {
		log.Panic(err)
	}
}
