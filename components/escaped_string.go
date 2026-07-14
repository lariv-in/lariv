package components

import (
	"context"

	. "maragu.dev/gomponents"
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

// Build returns the content wrapped in a Text Node
func (e EscapedString) Build(ctx context.Context) Node {
	return Text(e.Content)
}
