package components

import (
	"context"
	"fmt"
	"strings"

	"maragu.dev/gomponents"
	"maragu.dev/gomponents/html"
)

// TopbarButton defines a button rendered in the top navigation bar.
// These are registered globally via a registry in lago, and plugins can add their own.
type TopbarButton struct {
	UID           string // unique HTML ID for the button element
	Icon          string // heroicon name
	IconAlt       string // alternate icon (for toggling, e.g. sun/moon)
	IconCondition string // Alpine.js condition for showing the primary icon vs alt
	URL           Getter // lazily resolved URL (e.g. from route registry)
	Target        string // Turbo frame target selector
	Method        string // HTTP method: get (default), post, put, delete
	OnClick       string // JavaScript onclick handler
	Classes       string // CSS classes for the button
}

// TopbarButtonsGetter is a package-level getter that provides []TopbarButton.
// It is set by lago during initialization to read from RegistryTopbarButtons.
// LayoutTopbar reads from this automatically — no per-page configuration needed.
var TopbarButtonsGetter Getter

type LayoutTopbar struct {
	Page
	Children []PageInterface
}

func (e LayoutTopbar) Build(ctx context.Context) gomponents.Node {
	buttonNodes := gomponents.Group{}

	if TopbarButtonsGetter != nil {
		if buttons, ok := TopbarButtonsGetter(ctx).([]TopbarButton); ok {
			for _, btn := range buttons {
				// Resolve URL from getter
				url := ""
				if btn.URL != nil {
					if u, ok := btn.URL(ctx).(string); ok {
						url = u
					}
				}

				// Build icon node(s)
				var iconNode gomponents.Node
				if btn.IconAlt != "" && btn.IconCondition != "" {
					iconNode = gomponents.Group{Render(Icon{
						Name:  btn.Icon,
						Attrs: []gomponents.Node{gomponents.Attr("x-show", btn.IconCondition)},
					}, ctx), Render(Icon{
						Name:  btn.IconAlt,
						Attrs: []gomponents.Node{gomponents.Attr("x-show", fmt.Sprintf("!(%s)", btn.IconCondition))},
					}, ctx),
					}
				} else {
					iconNode = Render(Icon{Name: btn.Icon}, ctx)
				}

				// Collect button attributes
				attrs := []gomponents.Node{
					html.Class(fmt.Sprintf("btn %s", btn.Classes)),
					iconNode,
				}
				if btn.UID != "" {
					attrs = append(attrs, html.ID(btn.UID))
				}
				if btn.OnClick != "" {
					attrs = append(attrs, gomponents.Attr("onclick", btn.OnClick))
				}
				if url != "" {
					method := strings.ToLower(btn.Method)
					if method != "" && method != "get" {
						attrs = append(attrs, gomponents.Attr("hx-"+method, url))
					} else {
						attrs = append(attrs, html.Href(url))
					}
				}
				if btn.Target != "" {
					attrs = append(attrs, gomponents.Attr("hx-target", btn.Target))
				}

				if url != "" {
					buttonNodes = append(buttonNodes, html.A(attrs...))
				} else {
					buttonNodes = append(buttonNodes, html.Button(attrs...))
				}
			}
		}
	}

	childGroup := gomponents.Group{}
	for _, child := range e.Children {
		childGroup = append(childGroup, Render(child, ctx))
	}

	return html.Div(html.Class("h-screen flex flex-col overflow-hidden"),
		html.Div(html.Class("navbar bg-base-100 border-b border-base-300 px-4 flex justify-between items-center flex-none"),
			html.Div(html.Class("flex-1"),
				html.A(html.Href("/"), html.Class("text-xl font-bold"), gomponents.Text("Lago")),
			),
			html.Div(html.Class("flex-none flex items-center gap-2"),
				buttonNodes,
			),
		),
		html.Div(html.Class("flex-1 overflow-hidden"),
			childGroup,
		),
	)
}

func (e LayoutTopbar) GetChildren() []PageInterface {
	return e.Children
}
