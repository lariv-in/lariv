package components

import (
	"context"

	"github.com/lariv-in/components"
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

	return gomponents.El("details",
		html.Class("dropdown dropdown-end"),
		gomponents.Attr("@click.outside", "$el.removeAttribute('open')"),
		gomponents.El("summary",
			html.Class("btn btn-ghost btn-sm btn-circle avatar placeholder"),
			html.Div(
				html.Class("bg-neutral text-neutral-content rounded-full w-10"),
				html.Span(html.Class("text-xl"), gomponents.Text(avatarText)),
			),
		),
		html.Div(
			html.Class("card w-64 my-1.5 card-body shadow dropdown-content border border-base-300 rounded-box z-50 bg-base-100 p-4"),
			html.Div(
				html.Class("flex flex-col gap-1"),
				html.Div(html.Class("font-bold text-lg"), gomponents.Text(name)),
				html.Div(html.Class("text-sm opacity-70 cursor-default"), gomponents.Text(roleName)),
			),
		),
	)
}
