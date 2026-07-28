package components

import (
	"context"
	"html/template"
	"io"
	"log/slog"

	"github.com/lariv-in/lariv/getters"
)

// SidebarMenuItem represents a single clickable option hyperlink in a sidebar navigation list.
// It renders an HTML anchor inside a list element (`<li><a>`) featuring an optional icon and title.
//
// Use Cases:
//   - Defining navigation links inside sidebars or top navigation bar dropdown layouts.
//
// Example:
//
//	&components.SidebarMenuItem{
//	    Title:  getters.Static("Audit Log"),
//	    Url:    lariv.RoutePath("admin.AuditLogs", nil),
//	    Icon:   "document-text",
//	    Active: true,
//	}
type SidebarMenuItem struct {
	// Page embeds common component properties like Key and Roles.
	Page
	// Title is the dynamic function retrieving the display label text.
	Title getters.Getter[string]
	// Url is the dynamic function retrieving the destination anchor target path.
	Url getters.Getter[string]
	// Icon represents the SVG icon name representing this menu item.
	Icon string
	// Active specifies if this link is currently selected and highlighted.
	Active bool
}

// GetKey returns the unique key identifier for this SidebarMenuItem component.
func (e SidebarMenuItem) GetKey() string {
	return e.Key
}

// GetRoles returns the authorized roles required to view this SidebarMenuItem.
func (e SidebarMenuItem) GetRoles() []string {
	return e.Roles
}

// Build compiles the SidebarMenuItem component into a list item wrapping a navigation link.
func (e SidebarMenuItem) Build(cat Catalog, ctx context.Context, w io.Writer) error {
	title := ""
	if e.Title != nil {
		t, err := e.Title(ctx)
		if err != nil {
			slog.Error("SidebarMenuItem Title getter failed", "error", err, "key", e.Key)
			return ContainerError{Error: getters.Static(err)}.Build(cat, ctx, w)
		}
		title = t
	}
	url := "#"
	if e.Url != nil {
		u, err := e.Url(ctx)
		if err != nil {
			slog.Error("SidebarMenuItem Url getter failed", "error", err, "key", e.Key)
			return ContainerError{Error: getters.Static(err)}.Build(cat, ctx, w)
		}
		url = u
	}

	var iconHTML template.HTML
	if e.Icon != "" {
		h, err := RenderHTML(Icon{Name: e.Icon, Classes: "heroicon-sm"}, cat, ctx)
		if err != nil {
			return err
		}
		iconHTML = h
	}

	activeClass := ""
	if e.Active {
		activeClass = " menu-active"
	}

	return Execute(w, "sidebar_menu_item", struct {
		URL         string
		ActiveClass string
		Icon        template.HTML
		Title       string
	}{URL: url, ActiveClass: activeClass, Icon: iconHTML, Title: title})
}

// SidebarMenu represents a wrapper panel housing a collection of SidebarMenuItem list links.
// It bundles the list nodes in a DaisyUI menu wrapper (`<ul>`), supporting a header Category Title and an optional Back Button at the top.
//
// Use Cases:
//   - Defining navigation option categories or back-navigation controls within collapsible sidebar drawers.
//
// Example:
//
//	&components.SidebarMenu{
//	    Title: getters.Static("Management"),
//	    Children: []components.PageInterface{
//	        &components.SidebarMenuItem{
//	            Title: getters.Static("Settings"),
//	            Url:   lariv.RoutePath("admin.Settings", nil),
//	            Icon:  "cog",
//	        },
//	    },
//	}
type SidebarMenu struct {
	// Page embeds common component properties like Key and Roles.
	Page
	// Title represents the heading label text displayed at the top of the menu group.
	Title getters.Getter[string]
	// Back represents an optional back-navigation button displayed above the menu title.
	Back *SidebarMenuItem
	// Children represents the slice of sub-components rendered in the menu block list.
	Children []PageInterface
}

// Build compiles the SidebarMenu component into a list wrapper Ul element.
func (e SidebarMenu) Build(cat Catalog, ctx context.Context, w io.Writer) error {
	var backHTML template.HTML
	hasBack := false
	if e.Back != nil {
		backTitle := ""
		if e.Back.Title != nil {
			t, err := e.Back.Title(ctx)
			if err != nil {
				slog.Error("SidebarMenu Back Title getter failed", "error", err, "key", e.Key)
				return ContainerError{Error: getters.Static(err)}.Build(cat, ctx, w)
			}
			backTitle = t
		}
		backUrl := "#"
		if e.Back.Url != nil {
			u, err := e.Back.Url(ctx)
			if err != nil {
				slog.Error("SidebarMenu Back Url getter failed", "error", err, "key", e.Key)
				return ContainerError{Error: getters.Static(err)}.Build(cat, ctx, w)
			}
			backUrl = u
		}
		h, err := RenderHTML(ButtonLink{
			Page:    e.Back.Page,
			Label:   getters.Static(backTitle),
			Link:    getters.Static(backUrl),
			Icon:    "arrow-left",
			Classes: "btn-sm mb-2",
		}, cat, ctx)
		if err != nil {
			return err
		}
		backHTML = h
		hasBack = true
	}

	title := ""
	if e.Title != nil {
		t, err := e.Title(ctx)
		if err != nil {
			slog.Error("SidebarMenu Title getter failed", "error", err, "key", e.Key)
			return ContainerError{Error: getters.Static(err)}.Build(cat, ctx, w)
		}
		title = t
	}

	children, err := RenderChildren(cat, ctx, e.Children)
	if err != nil {
		return err
	}

	return Execute(w, "sidebar_menu", struct {
		HasBack  bool
		Back     template.HTML
		Title    string
		Children template.HTML
	}{HasBack: hasBack, Back: backHTML, Title: title, Children: children})
}

// GetKey returns the unique key identifier for this SidebarMenu component.
func (e SidebarMenu) GetKey() string {
	return e.Key
}

// GetRoles returns the authorized roles required to view this SidebarMenu.
func (e SidebarMenu) GetRoles() []string {
	return e.Roles
}

// GetChildren returns the slice of nested sub-components.
func (e SidebarMenu) GetChildren() []PageInterface {
	return e.Children
}

// SetChildren replaces the slice of nested sub-components.
func (e *SidebarMenu) SetChildren(children []PageInterface) {
	e.Children = children
}
