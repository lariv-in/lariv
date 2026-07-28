package components

import (
	"context"
	"fmt"
	"html/template"
	"io"
	"log/slog"

	"github.com/lariv-in/lariv/getters"
)

// ShowIf represents a conditional container component.
// It evaluates its dynamic Getter and compiles its Children within a wrapper Div only when the resolved value is truthy.
//
// Use Cases:
//   - Conditionally displaying actions, administrator credentials alerts, verification details, or toggle components based on state.
//
// Example:
//
//	&components.ShowIf{
//	    Getter: func(ctx context.Context) (any, error) {
//	        return user.IsAdmin, nil
//	    },
//	    Children: []components.PageInterface{
//	        &components.ButtonLink{Label: getters.Static("Admin console"), Link: lariv.RoutePath("admin.Console", nil)},
//	    },
//	}
type ShowIf struct {
	// Page embeds common component properties like Key and Roles.
	Page
	// Getter is the dynamic function evaluated to check truthiness.
	Getter getters.Getter[any]
	// Children represents the slice of sub-components compiled if truthy.
	Children []PageInterface
}

// Build compiles the ShowIf component, rendering nested children if the resolved Getter value is truthy.
func (e ShowIf) Build(cat Catalog, ctx context.Context, w io.Writer) error {
	if e.Getter == nil {
		return nil
	}
	v, err := e.Getter(ctx)
	if err != nil {
		slog.Error("ShowIf getter failed", "error", err, "key", e.Key)
		return ContainerError{Error: getters.Static(err)}.Build(cat, ctx, w)
	}
	if !isTruthy(v) {
		return nil
	}
	children, err := RenderChildren(cat, ctx, e.Children)
	if err != nil {
		return err
	}
	return Execute(w, "show_if", struct {
		Children template.HTML
	}{Children: children})
}

// GetKey returns the unique key identifier for this ShowIf component.
func (e ShowIf) GetKey() string {
	return e.Key
}

// GetRoles returns the authorized roles required to view this ShowIf.
func (e ShowIf) GetRoles() []string {
	return e.Roles
}

// GetChildren returns the slice of nested sub-components.
func (e ShowIf) GetChildren() []PageInterface {
	return e.Children
}

// SetChildren replaces the slice of nested sub-components.
func (e *ShowIf) SetChildren(children []PageInterface) {
	e.Children = children
}

// isTruthy determines if any interface value resolves to a logical true state.
func isTruthy(v any) bool {
	if v == nil {
		return false
	}
	switch t := v.(type) {
	case bool:
		return t
	case string:
		return t != ""
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		// This is so stupid
		return fmt.Sprintf("%d", t) != "0"
	default:
		return true
	}
}
