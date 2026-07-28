package components

import (
	"context"
	"fmt"
	"html/template"
	"io"
	"sort"

	"github.com/lariv-in/lariv/getters"
)

// ClientMatchIf evaluates a dynamic variable name and conditionally renders a matching component.
// It generates multiple Alpine.js x-if templates, essentially serving as a switch/case block on the client side.
//
// Use Cases:
//   - Showing different form fields depending on a dropdown selection (e.g., selecting "Email" vs "Phone" as a preferred contact method).
//   - Swapping out interactive widgets depending on user-selected views or tabs.
//
// Example:
//
//	&components.ClientMatchIf{
//	    Key: getters.Static("preferredContact"),
//	    Match: getters.Static(map[string]components.PageInterface{
//	        "email": &components.InputEmail{Label: "Email Address", Name: "email"},
//	        "phone": &components.InputPhone{Label: "Phone Number", Name: "phone"},
//	    }),
//	}
type ClientMatchIf struct {
	// Page embeds common component properties like Key and Roles.
	Page
	// Key is a Getter resolving to the name of the reactive client-side variable to evaluate.
	Key getters.Getter[string]
	// Match is a Getter resolving to a map where each key corresponds to a matching value of the variable,
	// and the value represents the component to render.
	Match getters.Getter[map[string]PageInterface]
	// Children represents the child components managed by this wrapper (typically unused directly in Build).
	Children []PageInterface
}

type clientMatchCase struct {
	Condition string
	Content   template.HTML
}

// Build compiles the ClientMatchIf component into a collection of conditional HTML <template> elements,
// one for each case inside the match map.
func (e ClientMatchIf) Build(cat Catalog, ctx context.Context, w io.Writer) error {
	if e.Key == nil {
		return nil
	}
	key, err := e.Key(ctx)
	if err != nil {
		return ContainerError{Error: getters.Static(err)}.Build(cat, ctx, w)
	}
	if e.Match == nil {
		return nil
	}
	match, err := e.Match(ctx)
	if err != nil {
		return ContainerError{Error: getters.Static(err)}.Build(cat, ctx, w)
	}
	keys := make([]string, 0, len(match))
	for k := range match {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	cases := make([]clientMatchCase, 0, len(keys))
	for _, k := range keys {
		page := match[k]
		if page == nil {
			continue
		}
		content, err := RenderHTML(page, cat, ctx)
		if err != nil {
			return err
		}
		cases = append(cases, clientMatchCase{
			Condition: fmt.Sprintf("%s === %q", key, k),
			Content:   content,
		})
	}
	return Execute(w, "client_match_if", struct {
		Cases []clientMatchCase
	}{Cases: cases})
}

// GetKey returns the unique key identifier for this ClientMatchIf component.
func (e ClientMatchIf) GetKey() string {
	return e.Page.Key
}

// GetRoles returns the authorized roles required to view this ClientMatchIf.
func (e ClientMatchIf) GetRoles() []string {
	return e.Roles
}

// GetChildren returns the slice of child components managed by this ClientMatchIf wrapper.
func (e ClientMatchIf) GetChildren() []PageInterface {
	return e.Children
}

// SetChildren overwrites the child components managed by this ClientMatchIf wrapper.
func (e *ClientMatchIf) SetChildren(children []PageInterface) {
	e.Children = children
}
