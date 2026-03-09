package components

import (
	"context"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type LayoutSidebar struct {
	Sidebar  []PageInterface
	Children []PageInterface
}

func (e LayoutSidebar) Build(ctx context.Context) Node {
	sidebarGroup := Group{}
	for _, child := range e.Sidebar {
		sidebarGroup = append(sidebarGroup, child.Build(ctx))
	}

	contentGroup := Group{}
	for _, child := range e.Children {
		contentGroup = append(contentGroup, child.Build(ctx))
	}

	return Div(ID("app-layout"), Class("size-full"),
		Attr("x-data", `{
        showLeft: window.innerWidth >= 768,
        isMobile: window.innerWidth < 768,
        messages: []
}`),
		Div(Class("grid h-full transition-[grid-template-columns] duration-[400ms] ease-in"),
			Attr(":class", "isMobile ? 'grid-cols-1' : (showLeft ? 'grid-cols-[250px_1fr]' : 'grid-cols-[0px_1fr]')"),

			// Mobile Overlay
			Div(
				Attr("x-show", "isMobile && showLeft"),
				Attr("x-transition.opacity", ""),
				Attr("@click", "showLeft = false"),
				Class("absolute inset-0 bg-black/50 z-20"),
			),

			// Sidebar
			Aside(
				Class("bg-base-100 border-r border-base-300 overflow-hidden"),
				Attr(":class", "isMobile ? 'absolute inset-y-0 left-0 z-50 shadow-xl transition-transform duration-300' : ''"),
				Attr(":style", "isMobile ? (showLeft ? 'transform: translateX(0)' : 'transform: translateX(-100%)') : ''"),
				Div(Class("h-full overflow-y-auto w-[250px] bg-base-100 p-2"),
					sidebarGroup,
				),
			),

			// Main Content
			Main(Class("overflow-y-auto p-4 relative h-full"),
				Button(
					Attr("@click", "showLeft = !showLeft"),
					Class("btn btn-sm btn-square mb-2"),
					Span(Class("heroicon heroicon-bars-3")),
				),

				// Messages (Simplified for now, will be populated via Alpine/HTMX)
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

func (e LayoutSidebar) GetChildren() []PageInterface {
	return append(e.Sidebar, e.Children...)
}
