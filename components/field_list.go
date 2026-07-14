package components

import (
	"context"
	"log/slog"

	"github.com/lariv-in/lago/getters"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// FieldList represents a generic read-only list component that iterates over a slice of items of type T.
// It resolves a slice from the context using the Getter, binds each item to the context key "$row" in turn,
// and renders the templates specified in the Children array for each item.
//
// Use Cases:
//   - Displaying nested sub-lists or itemized properties of a parent detail (e.g. phone numbers, email addresses, or related tag labels).
//
// Example:
//
//	&components.FieldList[PhoneNumber]{
//	    Getter: getters.Key[[]PhoneNumber]("$in.PhoneNumbers"),
//	    Children: []components.PageInterface{
//	        &components.FieldText{Getter: getters.Key[string]("$row.Number")},
//	    },
//	}
type FieldList[T any] struct {
	// Page embeds common component properties like Key and Roles.
	Page
	// Getter is the dynamic function retrieving the slice of items of type T.
	Getter getters.Getter[[]T] // resolves to a slice
	// Classes represents additional CSS classes applied to the output HTML outer div wrapper.
	// (Discouraged: Use layout containers or theme styling instead of custom styling overrides).
	Classes string
	// Children represents the template components rendered iteratively for each item row.
	Children []PageInterface // template for each item
}

// Build compiles the FieldList component by fetching the slice, setting the item in the context, and rendering the child template Nodes.
func (e FieldList[T]) Build(ctx context.Context) Node {
	var listNodes Group

	if e.Getter != nil {
		rawData, err := e.Getter(ctx)
		if err != nil {
			slog.Error("FieldList getter failed", "error", err, "key", e.Key)
			return ContainerError{Error: getters.Static(err)}.Build(ctx)
		}
		for _, item := range rawData {
			itemCtx := context.WithValue(ctx, "$row", item)
			var childrenNodes Group
			for _, child := range e.Children {
				childrenNodes = append(childrenNodes, Render(child, itemCtx))
			}
			listNodes = append(listNodes, Div(Class("list-item ml-4"), childrenNodes))
		}
	}

	return Div(Class(e.Classes), listNodes)
}

// GetKey returns the unique key identifier for this FieldList component.
func (e FieldList[T]) GetKey() string {
	return e.Key
}

// GetRoles returns the authorized roles required to view this FieldList.
func (e FieldList[T]) GetRoles() []string {
	return e.Roles
}

// GetChildren returns the slice of child components acting as templates in this list.
func (e FieldList[T]) GetChildren() []PageInterface {
	return e.Children
}

// SetChildren overwrites the child components acting as templates in this list.
func (e *FieldList[T]) SetChildren(children []PageInterface) {
	e.Children = children
}
