package components

import (
	"context"
	"io"
)

// EscapedString is a component that can be used to return escaped string
// as a component, useful when we just need to write some text
type EscapedString struct {
	// Page embeds common component properties like Key and Roles.
	Page
	// Content is the string that will be rendered
	Content string
}

// GetKey returns the unique key identifier for this EscapedString component.
func (e EscapedString) GetKey() string {
	return e.Key
}

// GetRoles returns the authorized roles required to view this EscapedString.
func (e EscapedString) GetRoles() []string {
	return e.Roles
}

// Build writes the escaped content.
func (e EscapedString) Build(cat Catalog, ctx context.Context, w io.Writer) error {
	return Execute(w, "escaped_string", struct{ Content string }{Content: e.Content})
}
