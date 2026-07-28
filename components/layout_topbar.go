package components

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"io"
	"strings"
)

// SidebarItem defines the icon and sub-component payload for drawer sections in the right sidebar.
type SidebarItem struct {
	// Icon represents the SVG icon name representing this sidebar tab button.
	Icon string
	// Content represents the child sub-component structure rendered in the sidebar pane.
	Content PageInterface
}

// LayoutTopbar represents a responsive page shell featuring a top navigation bar and a collapsible, resizable right sidebar drawer.
// Layout components are special structural nodes in Lariv establishing page wrappers. LayoutTopbar populates its navbar navigation items
// dynamically from [Catalog.TopbarItems], and populates right utility drawers dynamically from [Catalog.RightSidebarItems].
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

type layoutTopbarTab struct {
	ClassExpr string
	Click     string
	Icon      template.HTML
}

type layoutTopbarPanel struct {
	Show    string
	Content template.HTML
}

// Build compiles the LayoutTopbar component into a navbar and side container structure.
func (e LayoutTopbar) Build(cat Catalog, ctx context.Context, w io.Writer) error {
	if cat == nil {
		cat = EmptyCatalog{}
	}
	var topbarBuf bytes.Buffer
	for _, comp := range cat.TopbarItems() {
		if err := Render(comp, cat, ctx, &topbarBuf); err != nil {
			return err
		}
	}

	rightSidebarEntries := cat.RightSidebarItems()

	var xData string
	var keysJS string
	var defaultKey string
	hasSidebar := len(rightSidebarEntries) > 0
	if hasSidebar {
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

		showIcon, err := RenderHTML(Icon{
			Name:  "bars-3-bottom-right",
			Attrs: HTMLAttributes{"x-show": "!showRight"},
		}, cat, ctx)
		if err != nil {
			return err
		}
		hideIcon, err := RenderHTML(Icon{
			Name:  "x-mark",
			Attrs: HTMLAttributes{"x-show": "showRight"},
		}, cat, ctx)
		if err != nil {
			return err
		}
		if _, err := fmt.Fprintf(&topbarBuf,
			`<button class="btn btn-sm btn-square btn-ghost" @click="toggleRight()">%s%s</button>`,
			showIcon, hideIcon,
		); err != nil {
			return err
		}
	}

	children, err := RenderChildren(cat, ctx, e.Children)
	if err != nil {
		return err
	}

	var tabs []layoutTopbarTab
	var panels []layoutTopbarPanel
	showTabs := false
	if hasSidebar {
		showTabs = len(rightSidebarEntries) > 1
		if showTabs {
			for _, entry := range rightSidebarEntries {
				icon, err := RenderHTML(Icon{Name: entry.Item.Icon}, cat, ctx)
				if err != nil {
					return err
				}
				tabs = append(tabs, layoutTopbarTab{
					ClassExpr: fmt.Sprintf("activeTab === %q ? 'btn-primary' : 'btn-ghost'", entry.Key),
					Click:     fmt.Sprintf("setActiveTab(%q)", entry.Key),
					Icon:      icon,
				})
			}
		}
		for _, entry := range rightSidebarEntries {
			content, err := RenderHTML(entry.Item.Content, cat, ctx)
			if err != nil {
				return err
			}
			panels = append(panels, layoutTopbarPanel{
				Show:    fmt.Sprintf("activeTab === %q", entry.Key),
				Content: content,
			})
		}
	}

	return Execute(w, "layout_topbar", struct {
		XData        string
		TopbarItems  template.HTML
		HasSidebar   bool
		ShowTabs     bool
		Children     template.HTML
		Tabs         []layoutTopbarTab
		Panels       []layoutTopbarPanel
	}{
		XData:       xData,
		TopbarItems: template.HTML(topbarBuf.String()),
		HasSidebar:  hasSidebar,
		ShowTabs:    showTabs,
		Children:    children,
		Tabs:        tabs,
		Panels:      panels,
	})
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
