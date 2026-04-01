package components

import (
	"context"
	"log/slog"

	"github.com/lariv-in/lago/getters"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type SidebarMenuItem struct {
	Page
	Title  getters.Getter[string]
	Url    getters.Getter[string]
	Icon   string
	Active bool
}

func (e SidebarMenuItem) GetKey() string {
	return e.Key
}

func (e SidebarMenuItem) GetRoles() []string {
	return e.Roles
}

func (e SidebarMenuItem) Build(ctx context.Context) Node {
	title := ""
	if e.Title != nil {
		t, err := e.Title(ctx)
		if err != nil {
			slog.Error("SidebarMenuItem Title getter failed", "error", err, "key", e.Key)
			return ContainerError{Error: getters.Static(err)}.Build(ctx)
		}
		title = t
	}
	url := "#"
	if e.Url != nil {
		u, err := e.Url(ctx)
		if err != nil {
			slog.Error("SidebarMenuItem Url getter failed", "error", err, "key", e.Key)
			return ContainerError{Error: getters.Static(err)}.Build(ctx)
		}
		url = u
	}

	var iconNode Node
	if e.Icon != "" {
		iconNode = Render(Icon{Name: e.Icon, Classes: "heroicon-sm"}, ctx)
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
	Page
	Title    getters.Getter[string]
	Back     *SidebarMenuItem
	Children []PageInterface
}

func (e SidebarMenu) Build(ctx context.Context) Node {
	var items []Node

	// Back button
	if e.Back != nil {
		backTitle := ""
		if e.Back.Title != nil {
			t, err := e.Back.Title(ctx)
			if err != nil {
				slog.Error("SidebarMenu Back Title getter failed", "error", err, "key", e.Key)
				return ContainerError{Error: getters.Static(err)}.Build(ctx)
			}
			backTitle = t
		}
		backUrl := "#"
		if e.Back.Url != nil {
			u, err := e.Back.Url(ctx)
			if err != nil {
				slog.Error("SidebarMenu Back Url getter failed", "error", err, "key", e.Key)
				return ContainerError{Error: getters.Static(err)}.Build(ctx)
			}
			backUrl = u
		}
		items = append(items, Li(
			Render(ButtonLink{
				Page:    e.Back.Page,
				Label:   backTitle,
				Link:    getters.Static(backUrl),
				Icon:    "arrow-left",
				Classes: "btn-sm mb-2",
			}, ctx),
		))
	}

	// Title
	if e.Title != nil {
		title, err := e.Title(ctx)
		if err != nil {
			slog.Error("SidebarMenu Title getter failed", "error", err, "key", e.Key)
			return ContainerError{Error: getters.Static(err)}.Build(ctx)
		}
		if title != "" {
			items = append(items, Li(Class("menu-title font-semibold opacity-70"), Text(title)))
		}
	}

	// Children
	for _, child := range e.Children {
		items = append(items, Render(child, ctx))
	}

	return Ul(Class("menu w-full wrap-anywhere"), Group(items))
}

func (e SidebarMenu) GetKey() string {
	return e.Key
}

func (e SidebarMenu) GetRoles() []string {
	return e.Roles
}

func (e SidebarMenu) GetChildren() []PageInterface {
	return e.Children
}

func (e *SidebarMenu) SetChildren(children []PageInterface) {
	e.Children = children
}
