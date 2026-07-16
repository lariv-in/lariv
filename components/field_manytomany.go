package components

import (
	"context"
	"log/slog"

	"github.com/lariv-in/lariv/getters"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// FieldManyToMany represents a read-only layout displaying multiple associated records as tags or chips.
// It iterates through related records resolved by the Getter, utilizing the Display getter to print labels,
// and optionally links each chip to its detail resource URL using the Link getter.
//
// Use Cases:
//   - Showing list of taxes levied on an invoice line item.
//   - Displaying active system user roles or category tags on products.
//
// Example:
//
//	&components.FieldManyToMany[Tax]{
//	    Label:   "Applied Taxes",
//	    Getter:  getters.Key[[]Tax]("$in.Taxes"),
//	    Display: getters.Key[string]("$in.Name"),
//	    Link:    taxDetailURLGetter(),
//	}
type FieldManyToMany[T any] struct {
	// Page embeds common component properties like Key and Roles.
	Page
	// Label represents the header label text shown above the list of chips.
	Label string
	// Getter is the dynamic function retrieving the associated slices of type T.
	Getter getters.Getter[[]T]
	// Display is the Getter resolving the text label description string for each individual record.
	Display getters.Getter[string]
	// Link is an optional Getter resolving the detail navigation URL for each individual record.
	Link getters.Getter[string]
	// Classes represents additional CSS classes applied to the output HTML wrapper.
	// (Discouraged: Use layout containers or theme styling instead of custom styling overrides).
	Classes string
	// EmptyText is the fallback text message displayed if the Getter returns an empty slice (defaults to "—").
	EmptyText string
}

// GetKey returns the unique key identifier for this FieldManyToMany component.
func (e FieldManyToMany[T]) GetKey() string {
	return e.Key
}

// GetRoles returns the authorized roles required to view this FieldManyToMany.
func (e FieldManyToMany[T]) GetRoles() []string {
	return e.Roles
}

// Build compiles the FieldManyToMany component into an HTML panel rendering related selection chips.
func (e FieldManyToMany[T]) Build(ctx context.Context) Node {
	var chipNodes []Node
	if e.Getter != nil {
		values, err := e.Getter(ctx)
		if err != nil {
			slog.Error("FieldManyToMany getter failed", "error", err, "key", e.Key)
			return ContainerError{Error: getters.Static(err)}.Build(ctx)
		}
		for _, v := range values {
			pair, ok := manyToManySelectionPair(ctx, v, e.Display, e.Key)
			if !ok {
				continue
			}
			label := Span(Class("text-sm flex-1 min-w-0 truncate"), Text(pair.Value))
			chipClass := "flex items-center gap-1 rounded-lg bg-base-200 pl-2 pr-2 py-1 min-w-0 max-w-full"
			if e.Link != nil {
				itemCtx := context.WithValue(ctx, getters.ContextKeyIn, getters.MapFromStruct(v))
				href, err := e.Link(itemCtx)
				if err != nil {
					slog.Error("FieldManyToMany link getter failed", "error", err, "key", e.Key)
				} else if href != "" {
					chipNodes = append(chipNodes, A(
						Href(href),
						Class(chipClass+" link link-hover no-underline"),
						label,
					))
					continue
				}
			}
			chipNodes = append(chipNodes, Div(Class(chipClass), label))
		}
	}

	empty := e.EmptyText
	if empty == "" {
		empty = "—"
	}

	var body Node
	if len(chipNodes) == 0 {
		body = Span(Class("text-sm opacity-50"), Text(empty))
	} else {
		body = Div(
			Class("input input-bordered w-full min-h-12 h-auto flex flex-wrap items-center gap-2"),
			Group(chipNodes),
		)
	}

	outer := []Node{body}
	if e.Label != "" {
		outer = append([]Node{Label(Class("label text-sm font-bold"), Text(e.Label))}, outer...)
	}

	return Div(Class("my-1 "+e.Classes), Group(outer))
}
