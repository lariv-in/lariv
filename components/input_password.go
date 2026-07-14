package components

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/lariv-in/lago/getters"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// InputPassword represents a password or sensitive text input form field component.
// It renders an HTML password input (`<input type="password">`) masking user entries on the screen.
//
// Use Cases:
//   - Collecting passwords, secret credentials, access tokens, API secrets, or private keys.
//
// Example:
//
//	&components.InputPassword{
//	    Label:    "New Password",
//	    Name:     "new_password",
//	    Required: true,
//	}
type InputPassword struct {
	// Page embeds common component properties like Key and Roles.
	Page
	// Label represents the header label text displayed above the password input.
	Label string
	// Name represents the HTML form parameter name attribute.
	Name string
	// Getter is the dynamic function retrieving the default/current password string value (typically blank).
	Getter getters.Getter[string]
	// Required is a boolean indicating if this form password is a mandatory input.
	Required bool
	// Classes represents additional CSS classes applied to the output HTML wrapper.
	// (Discouraged: Use layout containers or theme styling instead of custom styling overrides).
	Classes string
}

// GetKey returns the unique key identifier for this InputPassword component.
func (e InputPassword) GetKey() string {
	return e.Key
}

// GetRoles returns the authorized roles required to view this InputPassword.
func (e InputPassword) GetRoles() []string {
	return e.Roles
}

// Build compiles the InputPassword component into a Div wrapping a password Input.
func (e InputPassword) Build(ctx context.Context) Node {
	valueNode := Value("")
	if e.Getter != nil {
		value, err := e.Getter(ctx)
		if err != nil {
			slog.Error("InputPassword getter failed", "error", err, "key", e.Key)
		} else {
			valueNode = Value(value)
		}
	}
	return Div(
		Class(fmt.Sprintf("my-1 %s", e.Classes)),
		Label(
			Class("label text-sm font-bold flex flex-col items-start gap-1"),
			Text(e.Label),
			Input(Type("password"), Name(e.Name), valueNode, Class(fmt.Sprintf("input input-bordered w-full %s", e.Classes)), If(e.Required, Required())),
		),
	)
}

// Parse extracts the password string value from input parameters.
func (e InputPassword) Parse(v any, _ context.Context) (any, error) {
	vals, _ := v.([]string)
	if len(vals) == 0 {
		return "", nil
	}
	// TODO: Add some password validation here
	return vals[0], nil
}

// GetName returns the HTML form element's name attribute value.
func (e InputPassword) GetName() string {
	return e.Name
}
