package components

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/lariv-in/lariv/getters"
	"github.com/lariv-in/lariv/registry"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// InputSelect represents a dropdown select menu input form field component.
// It renders an HTML select element populated with choices resolved dynamically from Choices.
// Parse verifies that the submitted option is in the choices list and returns key type T.
//
// Use Cases:
//   - Choosing choices from a defined list or enum (e.g. setting invoice status, choosing country regions, selecting system user roles).
//
// Example:
//
//	&components.InputSelect[string]{
//	    Label:   "Billing Status",
//	    Name:    "status",
//	    Choices: getters.Static([]registry.Pair[string, string]{{Key: "paid", Value: "Paid"}, {Key: "unpaid", Value: "Unpaid"}}),
//	    Getter:  getters.Static(registry.Pair[string, string]{Key: "unpaid", Value: "Unpaid"}),
//	}
type InputSelect[T comparable] struct {
	// Page embeds common component properties like Key and Roles.
	Page
	// Label represents the header label text displayed above the select element.
	Label string
	// Name represents the HTML form parameter name attribute.
	Name string
	// Choices is the dynamic function retrieving a slice of key-value choices of type registry.Pair[T, string].
	Choices getters.Getter[[]registry.Pair[T, string]]
	// Getter is the dynamic function retrieving the current/default selection of type registry.Pair[T, string].
	Getter getters.Getter[registry.Pair[T, string]]
	// Required is a boolean indicating if this form selection is mandatory.
	Required bool
	// EmptyLabel represents the custom label displayed for the empty value option if Required is false (defaults to "—").
	EmptyLabel string
	// Classes represents additional CSS classes applied to the output HTML wrapper.
	// (Discouraged: Use layout containers or theme styling instead of custom styling overrides).
	Classes string
	// Hidden specifies if this select field is rendered as a hidden input element.
	Hidden bool
	// Attr is an optional Getter returning additional HTML nodes/attributes to apply to the input.
	Attr getters.Getter[Node]
}

// GetKey returns the unique key identifier for this InputSelect component.
func (e InputSelect[T]) GetKey() string {
	return e.Key
}

// GetRoles returns the authorized roles required to view this InputSelect.
func (e InputSelect[T]) GetRoles() []string {
	return e.Roles
}

// Build compiles the InputSelect component into a Div wrapping a dropdown Select element.
func (e InputSelect[T]) Build(ctx context.Context) Node {
	var zero T

	choices := []registry.Pair[T, string]{}
	if e.Choices != nil {
		opts, err := e.Choices(ctx)
		if err != nil {
			slog.Error("InputSelect Choices getter failed", "error", err, "key", e.Key)
		} else {
			choices = opts
		}
	}

	rawSel := ""
	if e.Getter != nil {
		pair, err := e.Getter(ctx)
		if err != nil {
			slog.Error("InputSelect Getter failed", "error", err, "key", e.Key)
		} else if any(pair.Key) != any(zero) {
			rawSel = fmt.Sprint(pair.Key)
		}
	}

	emptyLab := "—"
	if e.EmptyLabel != "" {
		emptyLab = e.EmptyLabel
	}

	optionNodes := []Node{}
	if !e.Required {
		optionNodes = append(
			optionNodes,
			Option(Value(""), If(rawSel == "", Attr("selected", "")), Text(emptyLab)),
		)
	}
	for _, opt := range choices {
		ks := fmt.Sprint(opt.Key)
		optionNodes = append(
			optionNodes,
			Option(Value(ks), If(rawSel == ks, Attr("selected", "")), Text(opt.Value)),
		)
	}

	wrapClass := fmt.Sprintf("my-1 %s", e.Classes)
	if e.Hidden {
		wrapClass += " hidden"
	}
	return Div(
		Class(wrapClass),
		Label(
			Class("label text-sm font-bold flex flex-col items-start gap-1"),
			Text(e.Label),
			Select(
				Name(e.Name),
				Class(fmt.Sprintf("select select-bordered w-full %s", e.Classes)),
				Group(optionNodes),
				If(e.Required, Required()),
				Iff(e.Attr != nil, func() Node {
					n, err := e.Attr(ctx)
					if err != nil {
						slog.Error("InputSelect Attr getter failed", "error", err, "key", e.Key)
						return Raw("")
					}
					return n
				}),
			),
		),
	)
}

// Parse extracts and validates the selected value against the list of Choices, returning the key value.
func (e InputSelect[T]) Parse(v any, ctx context.Context) (any, error) {
	var zero T

	vals, ok := v.([]string)
	if !ok || len(vals) == 0 || vals[0] == "" {
		return zero, nil
	}
	submitted := vals[0]

	if e.Choices == nil {
		return zero, fmt.Errorf("InputSelect: no Choices getter for validation")
	}
	choices, err := e.Choices(ctx)
	if err != nil {
		return zero, fmt.Errorf("InputSelect: Choices getter failed: %w", err)
	}
	for _, opt := range choices {
		if fmt.Sprint(opt.Key) == submitted {
			return opt.Key, nil
		}
	}
	return zero, fmt.Errorf("InputSelect: invalid choice %q", submitted)
}

// GetName returns the HTML form element's name attribute value.
func (e InputSelect[T]) GetName() string {
	return e.Name
}
