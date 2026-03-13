package p_otp

import (
	"log"
	"net/url"

	"github.com/lariv-in/lago"
)

const AppURL = "/otp/preferences/"

func init() {
	u, err := url.Parse(AppURL)
	if err != nil {
		log.Panic(err)
	}

	err = lago.RegistryPlugins.Register("p_otp", lago.Plugin{
		Type:        lago.PluginTypeApp,
		Icon:        "key", // Using a relevant icon
		URL:         u,
		VerboseName: "OTP Preferences",
		RenderKeys:  []string{"superuser", "totschool_admin"},
	})
	if err != nil {
		log.Panic(err)
	}
}
