package p_otp

import (
	"log"
	"net/url"

	"github.com/lariv-in/lago/lago"
)

const AppURL = "/otp/preferences/"

func init() {
	u, err := url.Parse(AppURL)
	if err != nil {
		log.Panic(err)
	}

	err = lago.RegistryPlugin.Register("p_otp", lago.Plugin{
		Type:        lago.PluginTypeApp,
		Icon:        "key",
		URL:         u,
		VerboseName: "OTP Preferences",
		Roles:       []string{""},
	})
	if err != nil {
		log.Panic(err)
	}
}
