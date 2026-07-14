package components

import (
	"context"

	. "maragu.dev/gomponents"
)

// ClientIf conditionally renders its children on the client side using Alpine.js's x-if directive.
// It wraps the children within an HTML <template> element decorated with x-if.
//
// Use Cases:
//   - Toggling the visibility of a sub-section when a checkbox is ticked (e.g. showing a billing address form only when "Billing address is different" is checked).
//   - Disclosing detailed info segments when a user clicks a details button.
//
// Example:
//
//	 &components.ClientIf{
//	     Condition: "differentBillingAddress",
//	     Children: []components.PageInterface{
//	         &components.InputText{Label: "Billing Address", Name: "billing_address"},
//	     },
//	 }
type ClientIf struct {
	// Page embeds common component properties like Key and Roles.
	Page
	// Condition specifies a valid client-side Alpine.js expression to evaluate (e.g. "open").
	Condition string
	// Data is an optional Alpine.js data string (unused in the default Build method).
	Data string
	// Init is an optional Alpine.js initialization expression (unused in the default Build method).
	Init     string
	// Children represents the child components rendered conditionally.
	Children []PageInterface
}

// Build compiles the ClientIf component into an HTML <template> element.
// If multiple children are present, they are grouped under a single root <div> element to comply with Alpine's x-if requirements.
func (e ClientIf) Build(ctx context.Context) Node {
	group := Group{}
	for _, child := range e.Children {
		group = append(group, Render(child, ctx))
	}

	content := Node(group)
	if len(group) > 1 {
		// Alpine x-if template content must have a single root element.
		content = El("div", group)
	}

	return El("template",
		If(e.Condition != "", Attr("x-if", e.Condition)),
		content,
	)
}

// GetKey returns the unique key identifier for this ClientIf component.
func (e ClientIf) GetKey() string {
	return e.Key
}

// GetRoles returns the authorized roles required to view this ClientIf.
func (e ClientIf) GetRoles() []string {
	return e.Roles
}

// GetChildren returns the slice of child components managed by this ClientIf wrapper.
func (e ClientIf) GetChildren() []PageInterface {
	return e.Children
}

// SetChildren overwrites the child components managed by this ClientIf wrapper.
func (e *ClientIf) SetChildren(children []PageInterface) {
	e.Children = children
}
