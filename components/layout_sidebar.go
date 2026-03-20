package components

import (
	"context"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type LayoutSidebar struct {
	Page
	Sidebar  []PageInterface
	Children []PageInterface
}

func (e LayoutSidebar) Build(ctx context.Context) Node {
	sidebarGroup := Group{}
	for _, child := range e.Sidebar {
		sidebarGroup = append(sidebarGroup, Render(child, ctx))
	}

	contentGroup := Group{}
	for _, child := range e.Children {
		contentGroup = append(contentGroup, Render(child, ctx))
	}

	return Div(ID("app-layout"), Class("size-full"),
		Attr("x-data", `{
        showLeft: window.innerWidth >= 768,
        isMobile: window.innerWidth < 768,
        messages: []
}`),
		Div(Class("grid h-full transition-[grid-template-columns] duration-[400ms] ease-in"),
			Attr(":class", "isMobile ? 'grid-cols-1' : (showLeft ? 'grid-cols-[250px_1fr]' : 'grid-cols-[0px_1fr]')"),

			// Mobile Overlay (below topbar)
			Div(
				Attr("x-show", "isMobile && showLeft"),
				Attr("x-transition.opacity", ""),
				Attr("@click", "showLeft = false"),
				// top-16 matches the navbar height in LayoutTopbar
				Class("absolute inset-x-0 bottom-0 top-16 bg-black/50 z-20"),
			),

			// Sidebar
			Aside(
				Class("bg-base-100 border-r border-base-300 overflow-hidden"),
				Attr(":class", "isMobile ? 'absolute left-0 top-16 h-[calc(100vh-4rem)] z-50 shadow-xl transition-transform duration-300' : ''"),
				Attr(":style", "isMobile ? (showLeft ? 'transform: translateX(0)' : 'transform: translateX(-100%)') : ''"),
				Div(Class("h-full overflow-y-auto w-[250px] bg-base-100 p-2"),
					sidebarGroup,
				),
			),

			// Main Content
			Main(Class("overflow-y-auto p-4 relative h-full bg-base-200"),
				Button(
					Attr("@click", "showLeft = !showLeft"),
					Class("btn btn-sm btn-square mb-2"), Render(Icon{Name: "bars-3"}, ctx),
				),

				// Messages (Simplified for now, will be populated via Alpine/Turbo)
				Div(Class("messages mb-4"),
					Template(Attr("x-for", "msg in messages"),
						Div(Class("alert shadow-lg mb-2"),
							Attr(":class", "msg.tags == 'error' ? 'alert-error' : (msg.tags == 'success' ? 'alert-success' : 'alert-info')"),
							Div(Class("flex-1"),
								Span(Class("font-semibold"), Attr("x-text", "msg.tags.charAt(0).toUpperCase() + msg.tags.slice(1) + ':'")),
								Span(Attr("x-text", "msg.text")),
							),
						),
					),
				),

				contentGroup,
			),
		),
	)
}

func (e LayoutSidebar) GetKey() string {
	return e.Key
}

func (e LayoutSidebar) GetRoles() []string {
	return e.Roles
}

func (e LayoutSidebar) GetChildren() []PageInterface {
	return append(e.Sidebar, e.Children...)
}

func (e *LayoutSidebar) SetChildren(children []PageInterface) {
	offset := 0
	nSidebar := len(e.Sidebar)
	end := offset + nSidebar
	if end > len(children) {
		end = len(children)
	}
	e.Sidebar = children[offset:end]
	offset = end
	if offset >= len(children) {
		return
	}
	nContent := len(e.Children)
	end = offset + nContent
	if end > len(children) {
		end = len(children)
	}
	e.Children = children[offset:end]
	offset = end
	if offset < len(children) {
		e.Children = append(e.Children, children[offset:]...)
	}
}
