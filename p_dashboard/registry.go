package p_dashboard

import (
	"github.com/lariv-in/components"
	"github.com/lariv-in/lago"
)

var RegistryTopbarButton lago.Registry[components.TopBarButton] = lago.NewRegistry[components.TopBarButton]()

func init() {
	RegistryTopbarButton.Register("theme", components.TopBarButton{
		Icon:    "sun",
		OnClick: "toggleTheme()",
		Classes: "btn-outline",
	})
	RegistryTopbarButton.Register("apps", components.TopBarButton{
		Icon:    "squares-2x2",
		Url:     lago.GetterRoute("dashboard.AppsPage"),
		Classes: "btn-neutral",
	})
	RegistryTopbarButton.Register("logout", components.TopBarButton{
		Icon:    "arrow-right-start-on-rectangle",
		Url:     lago.GetterRoute("users.LogoutRoute"),
		Method:  "POST",
		Classes: "btn-error",
	})
}
