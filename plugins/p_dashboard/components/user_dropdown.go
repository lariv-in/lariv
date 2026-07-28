package components

import (
	"context"
	"html/template"
	"io"
	"log/slog"

	"github.com/lariv-in/lariv"
	"github.com/lariv-in/lariv/components"
	"github.com/lariv-in/lariv/getters"
	"github.com/lariv-in/lariv/plugins/p_users"
)

type UserDropdown struct {
	components.Page
}

func (e UserDropdown) GetKey() string {
	return e.Key
}

func (e UserDropdown) GetRoles() []string {
	return e.Roles
}

func (e UserDropdown) Build(cat components.Catalog, ctx context.Context, w io.Writer) error {
	user, userOK := p_users.UserFromContextOptional(ctx)
	roleName, roleOK := p_users.RoleFromContextOptional(ctx)

	var name, roleDisplay string
	if userOK {
		name = user.Name
	}
	if roleOK {
		roleDisplay = roleName
	}

	avatarText := "?"
	if len(name) > 0 {
		avatarText = string(name[0])
	}

	var selfDetailHref string
	var logoutButton template.HTML
	if userOK {
		href, err := getters.IfOr(lariv.RoutePath("p_users.SelfDetailRoute", nil), ctx, "")
		if err != nil {
			slog.Error("user dropdown: resolve self detail route", "error", err)
		}
		selfDetailHref = href

		logoutButton, err = components.RenderHTML(components.ButtonPost{
			Label:   "Logout",
			Icon:    "arrow-right-start-on-rectangle",
			URL:     lariv.RoutePath("p_users.LogoutRoute", nil),
			Classes: "btn btn-error justify-start w-full",
		}, cat, ctx)
		if err != nil {
			return err
		}
	}

	return execute(w, "user_dropdown", struct {
		AvatarText     string
		Name           string
		RoleDisplay    string
		UserOK         bool
		SelfDetailHref string
		LogoutButton   template.HTML
	}{
		AvatarText:     avatarText,
		Name:           name,
		RoleDisplay:    roleDisplay,
		UserOK:         userOK,
		SelfDetailHref: selfDetailHref,
		LogoutButton:   logoutButton,
	})
}
