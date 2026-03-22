package p_nirmancampus_website

import (
	"github.com/lariv-in/lago/lago"
)

func init() {
	lago.RegistryView.Register("nirmancampus_website.HomeView",
		lago.GetPageView("nirmancampus_website.HomePage"))
}
