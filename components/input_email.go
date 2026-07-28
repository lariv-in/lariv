package components

import (
	"context"
	"io"
	"log/slog"
	"net/mail"

	"github.com/lariv-in/lariv/getters"
)

// InputEmail represents an email address input form field component.
// It renders an HTML email picker input and validates input strings using Go's `net/mail.ParseAddress`.
//
// Use Cases:
//   - Capturing user contact email addresses during user registrations, newsletters, profile setup, or support workflows.
//
// Example:
//
//	&components.InputEmail{
//	    Label:  "Primary Email",
//	    Name:   "primary_email",
//	    Getter: getters.Key[string]("$in.Email"),
//	}
type InputEmail struct {
	// Page embeds common component properties like Key and Roles.
	Page
	// Label represents the header label text displayed above the email input.
	Label string
	// Name represents the HTML form parameter name attribute.
	Name string
	// Getter is the dynamic function retrieving the default/current email string value.
	Getter getters.Getter[string]
	// Required is a boolean indicating if this email field is a mandatory input.
	Required bool
	// Classes represents additional CSS classes applied to the output HTML wrapper.
	// (Discouraged: Use layout containers or theme styling instead of custom styling overrides).
	Classes string
}

// GetKey returns the unique key identifier for this InputEmail component.
func (e InputEmail) GetKey() string {
	return e.Key
}

// GetRoles returns the authorized roles required to view this InputEmail.
func (e InputEmail) GetRoles() []string {
	return e.Roles
}

// Build compiles the InputEmail component into a Div wrapping an email selection Input.
func (e InputEmail) Build(cat Catalog, ctx context.Context, w io.Writer) error {
	value := ""
	if e.Getter != nil {
		v, err := e.Getter(ctx)
		if err != nil {
			slog.Error("InputEmail getter failed", "error", err, "key", e.Key)
		} else {
			value = v
		}
	}
	return Execute(w, "input_email", struct {
		Classes  string
		Label    string
		Name     string
		Value    string
		Required bool
	}{
		Classes:  e.Classes,
		Label:    e.Label,
		Name:     e.Name,
		Value:    value,
		Required: e.Required,
	})
}

// Parse extracts the email string value and validates it using mail.ParseAddress.
func (e InputEmail) Parse(v any, _ context.Context) (any, error) {
	vals, _ := v.([]string)
	if len(vals) == 0 {
		return "", nil
	}
	address, err := mail.ParseAddress(vals[0])
	if err != nil {
		return nil, err
	}
	return address.Address, nil
}

// GetName returns the HTML form element's name attribute value.
func (e InputEmail) GetName() string {
	return e.Name
}
