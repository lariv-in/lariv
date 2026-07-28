package components

import (
	"context"
	"html/template"
	"io"
)

// RawString is a component that can be used to return un-escaped string
// as a component, useful when we just need to write some html
type RawString struct {
	// Page embeds common component properties like Key and Roles.
	Page
	// Content is the string that will be rendered
	Content string
}

// GetKey returns the unique key identifier for this RawString component.
func (e RawString) GetKey() string {
	return e.Key
}

// GetRoles returns the authorized roles required to view this RawString.
func (e RawString) GetRoles() []string {
	return e.Roles
}

// Build writes the unescaped HTML content.
func (e RawString) Build(cat Catalog, ctx context.Context, w io.Writer) error {
	return Execute(w, "raw_string", struct{ Content template.HTML }{Content: template.HTML(e.Content)})
}
