package components

import (
	"context"

	"github.com/lariv-in/lago/registry"
	"maragu.dev/gomponents"
	"maragu.dev/gomponents/html"
)

var RegistryTopbar = registry.NewRegistry[PageInterface]()

type LayoutTopbar struct {
	Page
	Children []PageInterface
}

func (e LayoutTopbar) Build(ctx context.Context) gomponents.Node {
	topbarItems := gomponents.Group{}

	for _, comp := range *RegistryTopbar.AllStable(registry.RegisterOrder[PageInterface]{}) {
		item := comp.Value
		topbarItems = append(topbarItems, Render(item, ctx))
	}

	childGroup := gomponents.Group{}
	for _, child := range e.Children {
		childGroup = append(childGroup, Render(child, ctx))
	}

	return html.Div(html.Class("h-screen flex flex-col overflow-hidden"),
		html.Div(html.Class("navbar bg-base-100 border-b border-base-300 px-4 flex justify-between items-center flex-none"),
			html.Div(html.Class("flex-1")),
			html.Div(html.Class("flex-none flex items-center gap-2"),
				topbarItems,
			),
		),
		html.Div(html.Class("flex-1 overflow-hidden"),
			childGroup,
		),
	)
}

func (e LayoutTopbar) GetKey() string {
	return e.Key
}

func (e LayoutTopbar) GetRoles() []string {
	return e.Roles
}

func (e LayoutTopbar) GetChildren() []PageInterface {
	return e.Children
}

func (e *LayoutTopbar) SetChildren(children []PageInterface) {
	e.Children = children
}
