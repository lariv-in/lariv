package components

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/lariv-in/lago/getters"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// InputTernary represents a three-state (Yes / No / Not Set) select input form field component.
// It renders an HTML select element and parses submitted option keys into boolean values or nil (representing a SQL NULL or unset state).
//
// Use Cases:
//   - Editing nullable boolean fields or three-state options (e.g., job application status: Yes/No/Pending review, remote work allowance: Allowed/Prohibited/Unspecified).
//
// Example:
//
//	&components.InputTernary{
//	    Label:      "Subscribed to Newsletter",
//	    Name:       "newsletter_status",
//	    TrueLabel:  "Opted In",
//	    FalseLabel: "Opted Out",
//	    NoneLabel:  "Unknown",
//	}
type InputTernary struct {
	// Page embeds common component properties like Key and Roles.
	Page
	// Label represents the header label text displayed above the select element.
	Label string
	// Name represents the HTML form parameter name attribute.
	Name string
	// Getter is the dynamic function retrieving the default/current boolean state.
	Getter getters.Getter[bool]
	// TrueLabel represents the display label for the positive state (defaults to "Yes").
	TrueLabel string
	// FalseLabel represents the display label for the negative state (defaults to "No").
	FalseLabel string
	// NoneLabel represents the display label for the unset state (defaults to "Not Set").
	NoneLabel string
	// Classes represents additional CSS classes applied to the output HTML wrapper.
	// (Discouraged: Use layout containers or theme styling instead of custom styling overrides).
	Classes string
}

// GetKey returns the unique key identifier for this InputTernary component.
func (e InputTernary) GetKey() string {
	return e.Key
}

// GetRoles returns the authorized roles required to view this InputTernary.
func (e InputTernary) GetRoles() []string {
	return e.Roles
}

// Build compiles the InputTernary component into a Div wrapping a three-state Select element.
func (e InputTernary) Build(ctx context.Context) Node {
	value := false
	hasValue := false
	if e.Getter != nil {
		v, err := e.Getter(ctx)
		if err != nil {
			slog.Error("InputTernary getter failed", "error", err, "key", e.Key)
		} else {
			value = v
			hasValue = true
		}
	}

	trueLabel := e.TrueLabel
	if trueLabel == "" {
		trueLabel = "Yes"
	}
	falseLabel := e.FalseLabel
	if falseLabel == "" {
		falseLabel = "No"
	}
	noneLabel := e.NoneLabel
	if noneLabel == "" {
		noneLabel = "Not Set"
	}

	noneSelected := ""
	trueSelected := ""
	falseSelected := ""
	if !hasValue {
		noneSelected = "selected"
	} else if value {
		trueSelected = "selected"
	} else {
		falseSelected = "selected"
	}

	return Div(
		Class(fmt.Sprintf("my-1 %s", e.Classes)),
		Label(
			Class("label text-sm font-bold flex flex-col items-start gap-1"),
			Text(e.Label),
			Select(
				Name(e.Name), Class("select select-bordered w-full"),
				Option(Value(""), If(noneSelected != "", Attr("selected", "")), Text(noneLabel)),
				Option(Value("True"), If(trueSelected != "", Attr("selected", "")), Text(trueLabel)),
				Option(Value("False"), If(falseSelected != "", Attr("selected", "")), Text(falseLabel)),
			),
		),
	)
}

// Parse extracts and parses selected string options into standard nullable Go boolean interfaces.
func (e InputTernary) Parse(v any, _ context.Context) (any, error) {
	vals, _ := v.([]string)
	if len(vals) == 0 || vals[0] == "" {
		return nil, nil
	}
	b, err := strconv.ParseBool(vals[0])
	if err != nil {
		return nil, nil
	}
	return b, nil
}

// GetName returns the HTML form element's name attribute value.
func (e InputTernary) GetName() string {
	return e.Name
}
