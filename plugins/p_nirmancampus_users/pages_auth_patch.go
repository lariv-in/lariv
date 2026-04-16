package p_nirmancampus_users

import (
	"log/slog"
	"net/http"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/views"
)

func init() {
	lago.RegistryPage.Patch("users.LoginPage", patchAuthShellRemoveSignupLinks)
	lago.RegistryPage.Patch("users.UnauthenticatedPage", patchAuthShellRemoveSignupLinks)

	lago.RegistryRoute.Patch("users.SignupRoute", func(old lago.Route) lago.Route {
		old.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			loginURL, err := lago.RoutePath("users.LoginRoute", nil)(r.Context())
			if err != nil {
				slog.Error("nirmancampus_users: resolve login route for signup redirect", "error", err)
				http.NotFound(w, r)
				return
			}
			views.HtmxRedirect(w, r, loginURL, http.StatusSeeOther)
		})
		return old
	})
}

// users.AuthSignupLink is defined on signup [components.ButtonLink] nodes in p_users (login + unauthenticated pages).
const authSignupLinkKey = "users.AuthSignupLink"

func patchAuthShellRemoveSignupLinks(page components.PageInterface) components.PageInterface {
	scaffold, ok := page.(*components.ShellAuthScaffold)
	if !ok {
		panic("nirmancampus_users: users.LoginPage / users.UnauthenticatedPage expected *ShellAuthScaffold")
	}
	if !components.RemoveChild[*components.ButtonLink](scaffold, authSignupLinkKey) {
		slog.Warn("nirmancampus_users: signup ButtonLink not found (expected key)", "key", authSignupLinkKey)
	}
	return scaffold
}
