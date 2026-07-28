package components

import (
	"context"
	"io"
	"log/slog"

	"github.com/lariv-in/lariv/getters"
	"github.com/nyaruka/phonenumbers"
)

// FieldPhone represents a read-only phone number display field.
// It parses the raw phone number value using the "IN" (India) region default and outputs it in E.164 international standard format.
//
// Use Cases:
//   - Showing normalized, standard-formatted phone contact details on user detail cards or contact pages.
//
// Example:
//
//	&components.FieldPhone{
//	    Getter:  getters.Key[string]("$in.Mobile"),
//	    Classes: "text-lg font-semibold",
//	}
type FieldPhone struct {
	// Page embeds common component properties like Key and Roles.
	Page
	// Getter is the dynamic function retrieving the raw phone number string.
	Getter getters.Getter[string]
	// Classes represents additional CSS classes applied to the output HTML div wrapper.
	// (Discouraged: Use layout containers or theme styling instead of custom styling overrides).
	Classes string
}

// GetKey returns the unique key identifier for this FieldPhone component.
func (e FieldPhone) GetKey() string {
	return e.Key
}

// GetRoles returns the authorized roles required to view this FieldPhone.
func (e FieldPhone) GetRoles() []string {
	return e.Roles
}

// Build compiles the FieldPhone component into a Div containing the E.164 formatted phone number.
func (e FieldPhone) Build(cat Catalog, ctx context.Context, w io.Writer) error {
	if e.Getter == nil {
		return nil
	}

	value, err := e.Getter(ctx)
	if err != nil {
		slog.Error("FieldPhone getter failed", "error", err, "key", e.Key)
		return ContainerError{Error: getters.Static(err)}.Build(cat, ctx, w)
	}

	v, err := phonenumbers.Parse(value, "IN")
	if err != nil {
		return ContainerError{Error: getters.Static(err)}.Build(cat, ctx, w)
	}

	return Execute(w, "field_phone", struct {
		Classes string
		Value   string
	}{Classes: e.Classes, Value: phonenumbers.Format(v, phonenumbers.E164)})
}
