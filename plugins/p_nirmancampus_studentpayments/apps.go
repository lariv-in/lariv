package p_nirmancampus_studentpayments

import (
	"log"
	"net/url"

	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_nirmancampus_students"
)

// AppUrl is under the Students app; see Caveats.md ("HTTP routes nested under another app's prefix").
var AppUrl = p_nirmancampus_students.AppUrl + "addon/studentpayments/"

func init() {
	u, err := url.Parse(AppUrl)
	if err != nil {
		log.Panic(err)
	}

	err = lago.RegistryPlugin.Register("p_nirmancampus_studentpayments", lago.Plugin{
		Type:        lago.PluginTypeAddon,
		Icon:        "banknote",
		URL:         u,
		VerboseName: "Student Payments",
		Roles:       []string{"superuser", "admin", "student"},
	})
	if err != nil {
		log.Panic(err)
	}
}
