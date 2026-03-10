package components

import (
	"context"
	"fmt"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type SidebarMenuItem struct {
	Title  Getter
	Url    Getter
	Icon   string
	Active bool
}

func (e SidebarMenuItem) Build(ctx context.Context) Node {
	title := fmt.Sprintf("%s", IfOrGetter(e.Title, ctx, ""))
	url := fmt.Sprintf("%s", IfOrGetter(e.Url, ctx, "#"))

	var iconNode Node
	if e.Icon != "" {
		iconNode = Icon{Name: e.Icon}.Build(ctx)
	}

	activeClass := ""
	if e.Active {
		activeClass = " menu-active"
	}

	return Li(
		A(Href(url), Class(activeClass),
			If(iconNode != nil, iconNode),
			Text(title),
		),
	)
}

type SidebarMenu struct {
	Title    Getter
	Back     *SidebarMenuItem
	Children []PageInterface
}

func (e SidebarMenu) Build(ctx context.Context) Node {
	var items []Node

	// Back button
	if e.Back != nil {
		backTitle := fmt.Sprintf("%s", IfOrGetter(e.Back.Title, ctx, ""))
		backUrl := fmt.Sprintf("%s", IfOrGetter(e.Back.Url, ctx, "#"))
		items = append(items, Li(
			A(Href(backUrl), Class("btn btn-sm mb-2"),
				Icon{Name: "arrow-left"}.Build(ctx),
				Text(backTitle),
			),
		))
	}

	// Title
	if e.Title != nil {
		title := fmt.Sprintf("%s", IfOrGetter(e.Title, ctx, ""))
		if title != "" {
			items = append(items, Li(Class("menu-title font-semibold opacity-70"), Text(title)))
		}
	}

	// Children
	for _, child := range e.Children {
		items = append(items, child.Build(ctx))
	}

	return Ul(Class("menu w-full wrap-anywhere"), Group(items))
}

func (e SidebarMenu) GetChildren() []PageInterface {
	return e.Children
}
