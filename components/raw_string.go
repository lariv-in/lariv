package components

import (
	"context"

	. "maragu.dev/gomponents"
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

// Build returns the content wrapped in a Raw Node
func (e RawString) Build(ctx context.Context) Node {
	return Raw(e.Content)
}
