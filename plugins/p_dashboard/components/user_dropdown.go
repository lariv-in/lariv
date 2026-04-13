package components

import (
	"context"
	"log/slog"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
	"maragu.dev/gomponents"
	"maragu.dev/gomponents/html"
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

func (e UserDropdown) Build(ctx context.Context) gomponents.Node {
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

	cardBody := gomponents.Group{
		html.Div(
			html.Class("flex flex-col gap-1"),
			html.Div(html.Class("font-bold text-lg"), gomponents.Text(name)),
			html.Div(html.Class("text-sm opacity-70 cursor-default"), gomponents.Text(roleDisplay)),
		),
	}
	if userOK {
		selfDetailHref, err := getters.IfOr(lago.RoutePath("users.SelfDetailRoute", nil), ctx, "")
		if err != nil {
			slog.Error("user dropdown: resolve self detail route", "error", err)
		}
		cardBody = append(cardBody, html.Div(
			html.Class("flex flex-col gap-1 mt-2 pt-2 border-t border-base-300"),
			html.A(
				html.Class("btn justify-start w-full"),
				html.Href(selfDetailHref),
				gomponents.Text("My Account"),
			),
			components.Render(components.ButtonPost{
				Label:   "Logout",
				Icon:    "arrow-right-start-on-rectangle",
				URL:     lago.RoutePath("users.LogoutRoute", nil),
				Classes: "btn btn-error justify-start w-full",
			}, ctx),
		))
	}

	return gomponents.El("details",
		html.Class("dropdown dropdown-end"),
		gomponents.Attr("@click.outside", "$el.removeAttribute('open')"),
		gomponents.El("summary",
			html.Class("btn btn-sm btn-circle avatar placeholder"),
			html.Div(
				html.Class("rounded-full w-10"),
				html.Span(html.Class("text-xl"), gomponents.Text(avatarText)),
			),
		),
		html.Div(
			html.Class("card w-64 my-1.5 card-body shadow dropdown-content border border-base-300 rounded-box z-50 bg-base-100 p-4"),
			cardBody,
		),
	)
}
