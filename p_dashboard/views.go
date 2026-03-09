package p_dashboard

import (
	"context"
	"net/http"

	"github.com/lariv-in/components"
	"github.com/lariv-in/lago"
	"github.com/lariv-in/p_users"
)

func renderTopbar(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buttonsMap := RegistryTopbarButton.All()
		var buttons []components.TopBarButton
		// We want a predictable order: theme, apps, logout
		keys := []string{"theme", "apps", "logout"}
		for _, k := range keys {
			if btn, ok := (*buttonsMap)[k]; ok {
				buttons = append(buttons, btn)
			}
		}
		// Also add any other registered buttons
		for k, btn := range *buttonsMap {
			isDefault := false
			for _, dk := range keys {
				if k == dk {
					isDefault = true
					break
				}
			}
			if !isDefault {
				buttons = append(buttons, btn)
			}
		}

		ctx := context.WithValue(r.Context(), "topbar_buttons", buttons)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func init() {
	lago.RegistryView.Register("dashboard.AppsView", p_users.AuthMiddleware(renderTopbar(lago.GetPageView("dashboard.AppsPage"))))
	err := lago.RegistryView.Patch("users.LoginSuccessView", func(_ http.Handler) http.Handler {
		return http.RedirectHandler("/apps/", http.StatusMovedPermanently)
	})
	if err != nil {
		panic(err)
	}
}
