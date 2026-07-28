package components

import (
	"context"
	"html/template"
	"io"

	"github.com/lariv-in/lariv/getters"
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
//	    URL: lariv.RoutePath("admin.ReportExportStatus", nil),
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

// Build compiles the HTMXPolling component into a Div with HTMX polling attributes.
func (e HTMXPolling) Build(cat Catalog, ctx context.Context, w io.Writer) error {
	children, err := RenderChildren(cat, ctx, e.Children)
	if err != nil {
		return err
	}
	url, err := e.URL(ctx)
	if err != nil {
		return ContainerError{Error: getters.Static(err)}.Build(cat, ctx, w)
	}
	return Execute(w, "htmx_polling", struct {
		URL      string
		Children template.HTML
	}{URL: url, Children: children})
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
