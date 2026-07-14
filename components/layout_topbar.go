package components

import (
	"context"
	"fmt"
	"strings"

	"github.com/lariv-in/lago/registry"
	"maragu.dev/gomponents"
	"maragu.dev/gomponents/html"
)

// RegistryTopbar holds registered sub-components (typically navigation menu items or user profile widgets)
// that are rendered dynamically within the top navigation bar.
var RegistryTopbar = registry.NewRegistry[PageInterface]()

// SidebarItem defines the icon and sub-component payload for drawer sections inside RegistryRightSidebar.
type SidebarItem struct {
	// Icon represents the SVG icon name representing this sidebar tab button.
	Icon string
	// Content represents the child sub-component structure rendered in the sidebar pane.
	Content PageInterface
}

// RegistryRightSidebar holds registered utility items to display in the layout's collapsible right drawer.
var RegistryRightSidebar = registry.NewRegistry[SidebarItem]()

// LayoutTopbar represents a responsive page shell featuring a top navigation bar and a collapsible, resizable right sidebar drawer.
// Layout components are special structural nodes in Lago establishing page wrappers. LayoutTopbar populates its navbar navigation items
// dynamically from RegistryTopbar, and populates right utility drawers dynamically from RegistryRightSidebar.
// The right sidebar features Alpine.js-driven click resizing, tab switching, and localStorage layout width persistence.
//
// Use Cases:
//   - Framing primary applications that feature persistent top menu navigations and secondary utility side panels (e.g., chat drawers, settings tabs, audit logs).
//
// Example:
//
//	&components.LayoutTopbar{
//	    Children: []components.PageInterface{
//	        &components.FieldTitle{Title: "Main Dashboard"},
//	    },
//	}
type LayoutTopbar struct {
	// Page embeds common component properties like Key and Roles.
	Page
	// Children represents the slice of sub-components rendered in the main layout viewport canvas.
	Children []PageInterface
}

// Build compiles the LayoutTopbar component into a navbar and side container structure.
func (e LayoutTopbar) Build(ctx context.Context) gomponents.Node {
	topbarItems := gomponents.Group{}

	for _, comp := range *RegistryTopbar.AllStable(registry.RegisterOrder[PageInterface]{}) {
		item := comp.Value
		topbarItems = append(topbarItems, Render(item, ctx))
	}

	// Fetch entries from RegistryRightSidebar
	rightSidebarEntries := *RegistryRightSidebar.AllStable(registry.RegisterOrder[SidebarItem]{})

	// Generate AlpineJS persistent state and control tags if there are sidebar items
	var xData string
	var keysJS string
	var defaultKey string
	if len(rightSidebarEntries) > 0 {
		defaultKey = rightSidebarEntries[0].Key

		var keysBuilder strings.Builder
		keysBuilder.WriteString("[")
		for i, entry := range rightSidebarEntries {
			if i > 0 {
				keysBuilder.WriteString(",")
			}
			keysBuilder.WriteString(fmt.Sprintf("%q", entry.Key))
		}
		keysBuilder.WriteString("]")
		keysJS = keysBuilder.String()

		xData = fmt.Sprintf(`{
			showRight: $persist(true).as('right-sidebar-show'),
			activeTab: $persist(%q).as('right-sidebar-active'),
			rightSidebarWidth: $persist(320).as('right-sidebar-width'),
			isResizing: false,
			init() {
				const keys = %s;
				if (!keys.includes(this.activeTab) && keys.length > 0) {
					this.activeTab = keys[0];
				}
			},
			toggleRight() {
				this.showRight = !this.showRight;
			},
			setActiveTab(key) {
				this.activeTab = key;
			},
			startResize(e) {
				e.preventDefault();
				this.isResizing = true;
				const startWidth = this.rightSidebarWidth;
				const startX = e.clientX;
				
				const onMouseMove = (moveEvent) => {
					if (!this.isResizing) return;
					const deltaX = moveEvent.clientX - startX;
					let newWidth = startWidth - deltaX;
					if (newWidth < 240) newWidth = 240;
					if (newWidth > 600) newWidth = 600;
					this.rightSidebarWidth = newWidth;
				};
				
				const onMouseUp = () => {
					this.isResizing = false;
					document.removeEventListener('mousemove', onMouseMove);
					document.removeEventListener('mouseup', onMouseUp);
				};
				
				document.addEventListener('mousemove', onMouseMove);
				document.addEventListener('mouseup', onMouseUp);
			}
		}`, defaultKey, keysJS)

		// Add toggle button to the topbar navigation menu if at least one item exists
		topbarItems = append(topbarItems, html.Button(
			html.Class("btn btn-sm btn-square btn-ghost"),
			gomponents.Attr("@click", "toggleRight()"),
			Render(Icon{
				Name:  "bars-3-bottom-right",
				Attrs: []gomponents.Node{gomponents.Attr("x-show", "!showRight")},
			}, ctx),
			Render(Icon{
				Name:  "x-mark",
				Attrs: []gomponents.Node{gomponents.Attr("x-show", "showRight")},
			}, ctx),
		))
	}

	childGroup := gomponents.Group{}
	for _, child := range e.Children {
		childGroup = append(childGroup, Render(child, ctx))
	}

	// Build the main layout
	var mainLayout gomponents.Node
	if len(rightSidebarEntries) > 0 {
		var asideAttrs []gomponents.Node
		asideAttrs = append(
			asideAttrs,
			html.Class("flex-none bg-base-100 flex flex-col h-full overflow-hidden absolute right-0 top-0 z-40 border-l border-base-300 shadow-2xl max-w-[85vw] sm:max-w-[400px] xl:static xl:border-l-0 xl:shadow-none xl:max-w-none"),
			gomponents.Attr("x-show", "showRight"),
			gomponents.Attr("x-transition:enter", "transition ease-out duration-200 transform"),
			gomponents.Attr("x-transition:enter-start", "translate-x-full"),
			gomponents.Attr("x-transition:enter-end", "translate-x-0"),
			gomponents.Attr("x-transition:leave", "transition ease-in duration-150 transform"),
			gomponents.Attr("x-transition:leave-start", "translate-x-0"),
			gomponents.Attr("x-transition:leave-end", "translate-x-full"),
			gomponents.Attr(":style", "'width: ' + rightSidebarWidth + 'px'"),
			gomponents.Attr("style", "width: 320px;"), // fallback default width
		)

		backdrop := html.Div(
			html.Class("xl:hidden absolute inset-0 bg-neutral-900/40 z-30 transition-opacity"),
			gomponents.Attr("x-show", "showRight"),
			gomponents.Attr("x-transition:enter", "transition ease-out duration-200"),
			gomponents.Attr("x-transition:enter-start", "opacity-0"),
			gomponents.Attr("x-transition:enter-end", "opacity-100"),
			gomponents.Attr("x-transition:leave", "transition ease-in duration-150"),
			gomponents.Attr("x-transition:leave-start", "opacity-100"),
			gomponents.Attr("x-transition:leave-end", "opacity-0"),
			gomponents.Attr("@click", "toggleRight()"),
		)

		resizer := html.Div(
			html.Class("hidden xl:flex w-2 -mx-1 cursor-col-resize flex-none h-full relative z-50 items-center justify-center hover:bg-primary/20 active:bg-primary/30 transition-all duration-150 group"),
			gomponents.Attr("x-show", "showRight"),
			gomponents.Attr("@mousedown", "startResize($event)"),
			gomponents.Attr(":class", "isResizing ? 'bg-primary/20' : ''"),
			html.Div(
				html.Class("w-[1px] h-full bg-base-300 group-hover:bg-primary group-active:bg-primary transition-colors duration-150"),
				gomponents.Attr(":class", "isResizing ? 'bg-primary' : ''"),
			),
		)

		// Tab Buttons Row (only if more than 1 item)
		var tabRow gomponents.Node = gomponents.Group{}
		if len(rightSidebarEntries) > 1 {
			var tabButtons []gomponents.Node
			for _, entry := range rightSidebarEntries {
				tabButtons = append(tabButtons, html.Button(
					html.Class("btn btn-sm btn-square"),
					gomponents.Attr(":class", fmt.Sprintf("activeTab === %q ? 'btn-primary' : 'btn-ghost'", entry.Key)),
					gomponents.Attr("@click", fmt.Sprintf("setActiveTab(%q)", entry.Key)),
					Render(Icon{Name: entry.Value.Icon}, ctx),
				))
			}
			tabRow = html.Div(
				html.Class("flex items-center gap-2 border-b border-base-300 p-2 overflow-x-auto flex-none"),
				gomponents.Group(tabButtons),
			)
		}

		// Content Panes
		var contentPanels []gomponents.Node
		for _, entry := range rightSidebarEntries {
			contentPanels = append(contentPanels, html.Div(
				gomponents.Attr("x-show", fmt.Sprintf("activeTab === %q", entry.Key)),
				html.Class("h-full overflow-y-auto p-0"),
				Render(entry.Value.Content, ctx),
			))
		}
		contentArea := html.Div(
			html.Class("flex-1 overflow-hidden relative"),
			gomponents.Group(contentPanels),
		)

		mainLayout = html.Div(
			html.Class("flex-1 flex overflow-hidden relative"),
			html.Div(
				html.Class("flex-1 overflow-hidden"),
				childGroup,
			),
			backdrop,
			resizer,
			html.Aside(
				append(
					asideAttrs,
					tabRow,
					contentArea,
				)...,
			),
		)
	} else {
		mainLayout = html.Div(
			html.Class("flex-1 overflow-hidden"),
			childGroup,
		)
	}

	rootAttrs := []gomponents.Node{
		html.Class("h-screen flex flex-col overflow-hidden"),
	}
	if len(rightSidebarEntries) > 0 {
		rootAttrs = append(rootAttrs, gomponents.Attr("x-data", xData))
	}
	rootAttrs = append(
		rootAttrs,
		html.Div(
			html.Class("navbar bg-base-100 border-b border-base-300 px-4 flex justify-between items-center flex-none"),
			html.Div(html.Class("flex-1")),
			html.Div(
				html.Class("flex-none flex items-center gap-2"),
				topbarItems,
			),
		),
		mainLayout,
	)

	return html.Div(rootAttrs...)
}

// GetKey returns the unique key identifier for this LayoutTopbar component.
func (e LayoutTopbar) GetKey() string {
	return e.Key
}

// GetRoles returns the authorized roles required to view this LayoutTopbar.
func (e LayoutTopbar) GetRoles() []string {
	return e.Roles
}

// GetChildren returns the slice of nested sub-components.
func (e LayoutTopbar) GetChildren() []PageInterface {
	return e.Children
}

// SetChildren replaces the slice of nested sub-components.
func (e *LayoutTopbar) SetChildren(children []PageInterface) {
	e.Children = children
}
