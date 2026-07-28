package components

import (
	"context"
	"io"
	"log/slog"
	"strings"

	"github.com/lariv-in/lariv/getters"
	"github.com/nyaruka/phonenumbers"
)

// InputPhone represents a telephone number input form field component.
// It renders an HTML tel input (`<input type="tel">`) and utilizes the phonenumbers library (default "IN" India region)
// to validate and format submitted numbers into the E.164 international standard.
//
// Use Cases:
//   - Collecting contact phone numbers for account profiles, checkout forms, support ticketing, or SMS OTP verifications.
//
// Example:
//
//	&components.InputPhone{
//	    Label:  "Primary Phone",
//	    Name:   "phone_number",
//	    Getter: getters.Key[string]("$in.PhoneNumber"),
//	}
type InputPhone struct {
	// Page embeds common component properties like Key and Roles.
	Page
	// Label represents the header label text displayed above the telephone input.
	Label string
	// Name represents the HTML form parameter name attribute.
	Name string
	// Getter is the dynamic function retrieving the default/current phone string value.
	Getter getters.Getter[string]
	// Required is a boolean indicating if this phone field is a mandatory input.
	Required bool
	// Classes represents additional CSS classes applied to the output HTML wrapper.
	// (Discouraged: Use layout containers or theme styling instead of custom styling overrides).
	Classes string
}

// GetKey returns the unique key identifier for this InputPhone component.
func (e InputPhone) GetKey() string {
	return e.Key
}

// GetRoles returns the authorized roles required to view this InputPhone.
func (e InputPhone) GetRoles() []string {
	return e.Roles
}

// Build compiles the InputPhone component into a Div wrapping a telephone Input.
func (e InputPhone) Build(cat Catalog, ctx context.Context, w io.Writer) error {
	displayValue := ""
	if e.Getter != nil {
		value, err := e.Getter(ctx)
		if err != nil {
			slog.Error("InputPhone getter failed", "error", err, "key", e.Key)
		} else {
			if value != "" {
				parsed, err := phonenumbers.Parse(value, "IN")
				if err == nil {
					displayValue = phonenumbers.Format(parsed, phonenumbers.E164)
				} else {
					displayValue = value
				}
			}
		}
	}
	return Execute(w, "input_phone", struct {
		Classes  string
		Label    string
		Name     string
		Value    string
		Required bool
	}{
		Classes:  e.Classes,
		Label:    e.Label,
		Name:     e.Name,
		Value:    displayValue,
		Required: e.Required,
	})
}

// Parse extracts text numbers from parameters and parses/formats them as E.164 phone formats.
func (e InputPhone) Parse(v any, _ context.Context) (any, error) {
	vals, _ := v.([]string)
	if len(vals) == 0 {
		return "", nil
	}
	raw := strings.TrimSpace(vals[0])
	if raw == "" {
		return "", nil
	}
	num, err := phonenumbers.Parse(raw, "IN")
	if err != nil {
		return nil, err
	}
	return phonenumbers.Format(num, phonenumbers.E164), nil
}

// GetName returns the HTML form element's name attribute value.
func (e InputPhone) GetName() string {
	return e.Name
}
