package components

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/lariv-in/lariv/getters"
	"github.com/lariv-in/lariv/registry"
	"gorm.io/datatypes"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// FieldKeyValue represents a read-only display field that renders dynamic key-value pairs.
// It parses a datatypes.JSON column containing a JSON array of [registry.Pair] elements (key/value strings)
// and displays them sequentially as key labels over text values.
//
// Use Cases:
//   - Showing dynamic attributes or metadata (e.g., custom attributes attached to products, server configurations, or transaction logs).
//
// Example:
//
//	&components.FieldKeyValue{
//	    Getter: getters.Key[datatypes.JSON]("$in.ConfigurationSettings"),
//	}
type FieldKeyValue struct {
	// Page embeds common component properties like Key and Roles.
	Page
	// Getter is the dynamic function retrieving the datatypes.JSON value containing the key-value list.
	Getter getters.Getter[datatypes.JSON]
	// Classes represents additional CSS classes applied to the output HTML div wrapper.
	// (Discouraged: Use layout containers or theme styling instead of custom styling overrides).
	Classes string
}

// GetKey returns the unique key identifier for this FieldKeyValue component.
func (e FieldKeyValue) GetKey() string {
	return e.Key
}

// GetRoles returns the authorized roles required to view this FieldKeyValue.
func (e FieldKeyValue) GetRoles() []string {
	return e.Roles
}

// Build compiles the FieldKeyValue component into an HTML structure rendering the parsed key-value lists.
func (e FieldKeyValue) Build(ctx context.Context) Node {
	if e.Getter == nil {
		return Div()
	}

	jsonData, err := e.Getter(ctx)
	if err != nil {
		slog.Error("FieldKeyValue getter failed", "error", err, "key", e.Key)
		return ContainerError{Error: getters.Static(err)}.Build(ctx)
	}
	if len(jsonData) == 0 {
		return Div()
	}

	var val []registry.Pair[string, string]
	err = json.Unmarshal(jsonData, &val)
	if err != nil {
		slog.Error("FieldKeyValue unmarshal failed", "error", err, "key", e.Key)
		return ContainerError{Error: getters.Static(err)}.Build(ctx)
	}

	var nodes []Node
	for _, r := range val {
		nodes = append(
			nodes,
			Div(
				Class("mb-4 pb-4 border-b border-base-300 last:border-b-0"),
				Div(Class("font-medium text-sm text-base-content/70 mb-1"), Text(r.Key)),
				Div(Class("whitespace-pre-wrap"), Text(r.Value)),
			),
		)
	}
	return Div(Class(e.Classes), Group(nodes))
}
