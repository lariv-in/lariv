package components

import (
	"context"

	"github.com/lariv-in/lago/getters"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// HTMXPolling represents a container component that periodically polls a server endpoint using HTMX.
// It sets up a trigger that sends a GET request to the resolved URL every 2 seconds (`hx-trigger="every 2s"`),
// replacing the entire page body (`hx-target="body"` and `hx-swap="outerHTML"`) with the returned response.
//
// Use Cases:
//   - Monitoring the progress of a long-running background job (e.g. tracking file export or data import jobs).
//   - Implementing auto-updating status screens or notification indicators.
//
// Example:
//
//	&components.HTMXPolling{
//	    URL: lago.RoutePath("admin.ReportExportStatus", nil),
//	    Children: []components.PageInterface{
//	        &components.FieldText{Getter: getters.Static("Document export in progress, please wait...")},
//	    },
//	}
type HTMXPolling struct {
	// Page embeds common component properties like Key and Roles.
	Page
	// URL is a Getter resolving to the target polling URL path.
	URL getters.Getter[string]
	// Children represents the nested components inside the polling div.
	Children []PageInterface
}

// Build compiles the HTMXPolling component into a Div Node with HTMX polling attributes.
func (e HTMXPolling) Build(ctx context.Context) Node {
	var children Group
	for _, child := range e.Children {
		children = append(children, Render(child, ctx))
	}
	url, err := e.URL(ctx)
	if err != nil {
		return ContainerError{Error: getters.Static(err)}.Build(ctx)
	}
	return Div(
		Attr("hx-get", url),
		Attr("hx-target", "body"),
		Attr("hx-swap", "outerHTML"),
		Attr("hx-trigger", "every 2s"),
		children,
	)
}

// GetKey returns the unique key identifier for this HTMXPolling component.
func (e HTMXPolling) GetKey() string {
	return e.Key
}

// GetRoles returns the authorized roles required to view this HTMXPolling.
func (e HTMXPolling) GetRoles() []string {
	return e.Roles
}

// GetChildren returns the slice of child components inside the polling container.
func (e HTMXPolling) GetChildren() []PageInterface {
	return e.Children
}

// SetChildren overwrites the child components inside the polling container.
func (e *HTMXPolling) SetChildren(children []PageInterface) {
	e.Children = children
}
