package components

import (
	"context"
	"log/slog"

	"github.com/lariv-in/components"
	"github.com/lariv-in/getters"
	"github.com/lariv-in/lago"
	"github.com/lariv-in/p_users"
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
	userObj := ctx.Value("$user")
	roleObj := ctx.Value("$role")

	var name, roleName string
	if user, ok := userObj.(p_users.User); ok {
		name = user.Name
	}
	if role, ok := roleObj.(string); ok {
		roleName = role
	}

	avatarText := "?"
	if len(name) > 0 {
		avatarText = string(name[0])
	}

	cardBody := gomponents.Group{
		html.Div(
			html.Class("flex flex-col gap-1"),
			html.Div(html.Class("font-bold text-lg"), gomponents.Text(name)),
			html.Div(html.Class("text-sm opacity-70 cursor-default"), gomponents.Text(roleName)),
		),
	}
	if _, ok := userObj.(p_users.User); ok {
		selfDetailHref, err := getters.IfOrGetter(lago.GetterRoutePath("users.SelfDetailRoute", nil), ctx, "")
		if err != nil {
			slog.Error("user dropdown: resolve self detail route", "error", err)
		}
		selfUpdateHref, err := getters.IfOrGetter(lago.GetterRoutePath("users.SelfUpdateRoute", nil), ctx, "")
		if err != nil {
			slog.Error("user dropdown: resolve self update route", "error", err)
		}
		cardBody = append(cardBody, html.Div(
			html.Class("flex flex-col gap-1 mt-2 pt-2 border-t border-base-300"),
			html.A(
				html.Class("btn btn-sm btn-ghost justify-start"),
				html.Href(selfDetailHref),
				gomponents.Text("My profile"),
			),
			html.A(
				html.Class("btn btn-sm btn-ghost justify-start"),
				html.Href(selfUpdateHref),
				gomponents.Text("Edit profile"),
			),
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
