package components

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/lariv-in/lago/fields"
	"github.com/lariv-in/lago/getters"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// InputPointsDecimal represents a high-precision decimal value input form field component.
// It parses decimal strings into custom [fields.DecimalSix] objects to avoid type-decoding mismatches during CRUD maps parsing.
//
// Use Cases:
//   - Inputting financial currencies, precise product weights, or fraction values (e.g. interest rates, currency conversion ratios).
//
// Example:
//
//	&components.InputPointsDecimal{
//	    Label:  "Interest Rate",
//	    Name:   "interest_rate",
//	    Getter: getters.Key[fields.DecimalSix]("$in.InterestRate"),
//	}
type InputPointsDecimal struct {
	// Page embeds common component properties like Key and Roles.
	Page
	// Label represents the header label text displayed above the decimal input.
	Label string
	// Name represents the HTML form parameter name attribute.
	Name string
	// Getter is the dynamic function retrieving the default/current fields.DecimalSix value.
	Getter getters.Getter[fields.DecimalSix]
	// Required is a boolean indicating if this form decimal is a mandatory input.
	Required bool
	// Classes represents additional CSS classes applied to the output HTML wrapper.
	// (Discouraged: Use layout containers or theme styling instead of custom styling overrides).
	Classes string
	// Hidden specifies if this decimal field is rendered as a hidden input element.
	Hidden bool
}

// GetKey returns the unique key identifier for this InputPointsDecimal component.
func (e InputPointsDecimal) GetKey() string { return e.Key }

// GetRoles returns the authorized roles required to view this InputPointsDecimal.
func (e InputPointsDecimal) GetRoles() []string { return e.Roles }

// Build compiles the InputPointsDecimal component into a Div wrapping a decimal text Input.
func (e InputPointsDecimal) Build(ctx context.Context) Node {
	text := ""
	if e.Getter != nil {
		pd, err := e.Getter(ctx)
		if err != nil {
			slog.Error("InputPointsDecimal getter failed", "error", err, "key", e.Key)
		} else {
			text = pd.String()
		}
	}
	wrapClass := fmt.Sprintf("my-1 %s", e.Classes)
	if e.Hidden {
		wrapClass += " hidden"
	}
	valueNode := Value(text)
	return Div(
		Class(wrapClass),
		Label(
			Class("label text-sm font-bold flex flex-col items-start gap-1"),
			If(!e.Hidden, Text(e.Label)),
			Input(
				If(!e.Hidden, Type("text")), If(e.Hidden, Type("hidden")), Name(e.Name),
				valueNode,
				Attr("inputmode", "decimal"),
				Class(fmt.Sprintf("input input-bordered w-full %s", e.Classes)),
				If(e.Required, Required()),
			),
		),
	)
}

// Parse extracts and unmarshals string values into a fields.DecimalSix object.
func (e InputPointsDecimal) Parse(v any, _ context.Context) (any, error) {
	vals, _ := v.([]string)
	if len(vals) == 0 || strings.TrimSpace(vals[0]) == "" {
		var out fields.DecimalSix
		if err := out.UnmarshalText([]byte("")); err != nil {
			return fields.DecimalSix{}, err
		}
		return out, nil
	}
	var out fields.DecimalSix
	if err := out.UnmarshalText([]byte(strings.TrimSpace(vals[0]))); err != nil {
		return fields.DecimalSix{}, err
	}
	return out, nil
}

// GetName returns the HTML form element's name attribute value.
func (e InputPointsDecimal) GetName() string { return e.Name }
