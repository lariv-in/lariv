package p_dashboard

import (

	"github.com/lariv-in/components"
	"github.com/lariv-in/lago"
)

func init() {
	components.RegistryTopbarButtons.Register("dashboard.themeButton", components.TopbarButton{
		UID:           "topbar-theme-btn",
		Icon:          "sun",
		IconAlt:       "moon",
		IconCondition: "theme === 'light'",
		OnClick:       "toggleTheme()",
		Classes:       "btn-sm btn-square btn-outline",
	})
	components.RegistryTopbarButtons.Register("dashboard.appsPageButton", components.TopbarButton{
		UID:     "topbar-apps-btn",
		Icon:    "squares-2x2",
		URL:     lago.GetterRoutePath("dashboard.AppsPage", nil),
		Classes: "btn-sm btn-square btn-neutral",
	})
	components.RegistryTopbarButtons.Register("dashboard.logoutButton", components.TopbarButton{
		UID:     "topbar-logout-btn",
		Icon:    "arrow-right-start-on-rectangle",
		URL:     lago.GetterRoutePath("users.LogoutRoute", nil),
		Method:  "post",
		Classes: "btn-sm btn-square btn-error",
	})
}
