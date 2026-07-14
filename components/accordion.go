// Package components provides a set of reusable UI components and input controls
// built on top of gomponents for structured, server-side rendered HTML generation.
package components

import (
	"context"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// AccordionItem represents a single collapsible section within an Accordion.
type AccordionItem struct {
	// Page embeds common component properties like Key and Roles.
	Page
	// Title is the header element/component displayed for the accordion item.
	Title PageInterface
	// Open determines if the accordion item is expanded by default.
	Open bool
	// Children are the content elements/components shown when the item is expanded.
	Children []PageInterface
}

// Accordion represents a collapsible group of panels, where each panel is an AccordionItem.
// It renders as a vertical stack of collapsible items.
type Accordion struct {
	// Page embeds common component properties like Key and Roles.
	Page
	// Classes specifies optional CSS classes to be appended to the root container.
	Classes string
	// Items is the list of collapsible sections within this accordion.
	Items []AccordionItem
}

// Build compiles the Accordion component into a gomponents Node representing the HTML structure.
func (e Accordion) Build(ctx context.Context) Node {
	var nodes []Node
	for _, item := range e.Items {
		var childNodes []Node
		for _, child := range item.Children {
			childNodes = append(childNodes, Render(child, ctx))
		}
		nodes = append(nodes,
			Div(Class("collapse collapse-arrow bg-base-100 border border-base-300"),
				El("input", Type("checkbox"), If(item.Open, Attr("checked", "checked"))),
				Div(Class("collapse-title"), Render(item.Title, ctx)),
				Div(Class("collapse-content"), Group(childNodes)),
			),
		)
	}
	return Div(Class("join join-vertical w-full "+e.Classes), Group(nodes))
}

// GetKey returns the unique key identifier for this Accordion.
func (e Accordion) GetKey() string {
	return e.Key
}

// GetRoles returns the authorized roles required to view this Accordion.
func (e Accordion) GetRoles() []string {
	return e.Roles
}

// GetChildren aggregates and returns all child elements/components from all AccordionItems.
func (e Accordion) GetChildren() []PageInterface {
	var all []PageInterface
	for _, item := range e.Items {
		all = append(all, item.Children...)
	}
	return all
}

// SetChildren distributes a flat slice of child components back across the AccordionItems.
// If there are more children than the items originally held, the remaining children are
// appended to the final item.
func (e *Accordion) SetChildren(children []PageInterface) {
	offset := 0
	for i := range e.Items {
		n := len(e.Items[i].Children)
		end := min(offset+n, len(children))
		e.Items[i].Children = children[offset:end]
		offset = end
		if offset >= len(children) {
			return
		}
	}
	if offset < len(children) && len(e.Items) > 0 {
		e.Items[len(e.Items)-1].Children = append(e.Items[len(e.Items)-1].Children, children[offset:]...)
	}
}
