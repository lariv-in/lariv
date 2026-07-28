package components

import (
	"context"
	"io"
	"log/slog"

	"github.com/lariv-in/lariv/getters"
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

type fieldManyToManyChip struct {
	Href  string
	Label string
}

// Build compiles the FieldManyToMany component into an HTML panel rendering related selection chips.
func (e FieldManyToMany[T]) Build(cat Catalog, ctx context.Context, w io.Writer) error {
	var chips []fieldManyToManyChip
	if e.Getter != nil {
		values, err := e.Getter(ctx)
		if err != nil {
			slog.Error("FieldManyToMany getter failed", "error", err, "key", e.Key)
			return ContainerError{Error: getters.Static(err)}.Build(cat, ctx, w)
		}
		for _, v := range values {
			pair, ok := manyToManySelectionPair(ctx, v, e.Display, e.Key)
			if !ok {
				continue
			}
			chip := fieldManyToManyChip{Label: pair.Value}
			if e.Link != nil {
				itemCtx := context.WithValue(ctx, getters.ContextKeyIn, getters.MapFromStruct(v))
				href, err := e.Link(itemCtx)
				if err != nil {
					slog.Error("FieldManyToMany link getter failed", "error", err, "key", e.Key)
				} else if href != "" {
					chip.Href = href
				}
			}
			chips = append(chips, chip)
		}
	}

	empty := e.EmptyText
	if empty == "" {
		empty = "—"
	}

	return Execute(w, "field_many_to_many", struct {
		Classes   string
		Label     string
		Empty     bool
		EmptyText string
		Chips     []fieldManyToManyChip
	}{
		Classes:   "my-1 " + e.Classes,
		Label:     e.Label,
		Empty:     len(chips) == 0,
		EmptyText: empty,
		Chips:     chips,
	})
}
