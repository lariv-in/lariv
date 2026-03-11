package p_dashboard

import (
	"context"

	"github.com/lariv-in/components"
	"github.com/lariv-in/lago"
)

func init() {
	prev := components.TopbarButtonsGetter
	components.TopbarButtonsGetter = func(ctx context.Context) any {
		buttons := []components.TopbarButton{}
		if prev != nil {
			if existing, ok := prev(ctx).([]components.TopbarButton); ok {
				buttons = append(buttons, existing...)
			}
		}
		return append(buttons,
			components.TopbarButton{
				UID:           "topbar-theme-btn",
				Icon:          "sun",
				IconAlt:       "moon",
				IconCondition: "theme === 'light'",
				OnClick:       "toggleTheme()",
				Classes:       "btn-sm btn-square btn-outline",
			},
			components.TopbarButton{
				UID:     "topbar-apps-btn",
				Icon:    "squares-2x2",
				URL:     lago.RoutePathGetter("dashboard.AppsPage"),
				Classes: "btn-sm btn-square btn-neutral",
			},
			components.TopbarButton{
				UID:     "topbar-logout-btn",
				Icon:    "arrow-right-start-on-rectangle",
				URL:     lago.RoutePathGetter("users.LogoutRoute"),
				Method:  "post",
				Classes: "btn-sm btn-square btn-error",
			},
		)
	}
}
