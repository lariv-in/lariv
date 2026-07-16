package components

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/url"
	"reflect"
	"strconv"
	"strings"

	"github.com/lariv-in/lariv/getters"
	"gorm.io/gorm"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// InputForeignKey represents a relationship selector form input component.
// It displays a clickable selection field that triggers an HTMX modal loaded from Url to present a list of choices (typically a table view).
// When an option is clicked, it bubbles an Alpine.js `@fk-select` event to populate the hidden input value and display name.
// During form submissions, Parse fetches the target record by ID from the database using GORM to validate its existence.
//
// Use Cases:
//   - Associating entities (e.g., selecting a customer for an invoice, assigning a department to a user, choosing a category for a product).
//
// Example:
//
//	&components.InputForeignKey[Department]{
//	    Label:       "User Department",
//	    Name:        "department_id",
//	    Getter:      getters.Key[Department]("$in.Department"),
//	    Display:     getters.Key[string]("$in.Name"),
//	    Url:         lariv.RoutePath("departments.SelectModal", nil),
//	}
type InputForeignKey[T any] struct {
	// Page embeds common component properties like Key and Roles.
	Page
	// Label represents the header label text displayed above the selector.
	Label string
	// Name represents the HTML form parameter name attribute.
	Name string
	// Getter is the dynamic function retrieving the currently selected model of type T.
	Getter getters.Getter[T]
	// Display is the Getter resolving the display text string from the selected model context.
	Display getters.Getter[string]
	// Placeholder represents the default text shown when no option is selected (defaults to "Select...").
	Placeholder string
	// Url is a Getter resolving the AJAX endpoint of the selection modal.
	Url getters.Getter[string]
	// Required is a boolean indicating if this form selection is mandatory.
	Required bool
	// Classes represents additional CSS classes applied to the output HTML wrapper.
	// (Discouraged: Use layout containers or theme styling instead of custom styling overrides).
	Classes string
	// Attr is an optional Getter returning additional HTML nodes/attributes to apply to the input.
	Attr getters.Getter[Node]
	// Hidden specifies if this selection field renders only a hidden input without a visible label or dialog trigger.
	Hidden bool
}

// GetKey returns the unique key identifier for this InputForeignKey component.
func (e InputForeignKey[T]) GetKey() string {
	return e.Key
}

// GetRoles returns the authorized roles required to view this InputForeignKey.
func (e InputForeignKey[T]) GetRoles() []string {
	return e.Roles
}

// Build compiles the InputForeignKey component into an Alpine-driven picker container Div.
func (e InputForeignKey[T]) Build(ctx context.Context) Node {
	valuePk := ""
	displayValue := ""

	if e.Getter != nil {
		value, err := e.Getter(ctx)
		if err != nil {
			slog.Error("InputForeignKey getter failed", "error", err, "key", e.Key)
		} else {
			valueMap := getters.MapFromStruct(value)
			if len(valueMap) > 0 {
				haveSelectedID := false
				if idVal, exists := valueMap["ID"]; exists {
					if rv := reflect.ValueOf(idVal); rv.IsValid() && !rv.IsZero() {
						valuePk = fmt.Sprintf("%v", idVal)
						haveSelectedID = true
					}
				} else if idVal, exists := valueMap["id"]; exists {
					if rv := reflect.ValueOf(idVal); rv.IsValid() && !rv.IsZero() {
						valuePk = fmt.Sprintf("%v", idVal)
						haveSelectedID = true
					}
				}
				if e.Display != nil && haveSelectedID {
					displayStr, err := e.Display(context.WithValue(ctx, "$in", valueMap))
					if err != nil {
						slog.Error("InputForeignKey display getter failed", "error", err, "key", e.Key)
					} else {
						displayValue = displayStr
					}
				}
			}
		}
	}

	if e.Hidden {
		wrapClass := fmt.Sprintf("my-1 %s", e.Classes)
		wrapClass += " hidden"
		return Div(
			Class(wrapClass),
			Input(
				Type("hidden"), Name(e.Name), Value(valuePk),
				Iff(e.Attr != nil, func() (out Node) {
					out = Raw("")
					n, err := e.Attr(ctx)
					if err != nil {
						slog.Error("InputForeignKey attr getter failed", "error", err, "key", e.Key)
						return out
					}
					return n
				}),
			),
		)
	}

	placeholder := e.Placeholder
	if placeholder == "" {
		placeholder = "Select..."
	}

	urlStr := ""
	if e.Url != nil {
		var err error
		urlStr, err = e.Url(ctx)
		if err != nil {
			slog.Error("InputForeignKey url getter failed", "error", err, "key", e.Key)
			urlStr = ""
		} else if urlStr != "" && e.Name != "" {
			if parsedURL, err := url.Parse(urlStr); err == nil {
				q := parsedURL.Query()
				q.Set("target_input", e.Name)
				parsedURL.RawQuery = q.Encode()
				urlStr = parsedURL.String()
			}
		}
	}

	alpinePayload, errAlpine := json.Marshal(map[string]string{
		"value":       valuePk,
		"display":     displayValue,
		"placeholder": placeholder,
	})
	if errAlpine != nil {
		alpinePayload = []byte(`{"value":"","display":"","placeholder":""}`)
	}
	alpineData := string(alpinePayload)
	// Selector dialog is closed from the table row @click (getters.Select); avoid removing the wrong dialog here.
	eventHandler := fmt.Sprintf("if ($event.detail.name === '%s') { value = $event.detail.value; display = $event.detail.display }", e.Name)

	return Div(
		Class(fmt.Sprintf("my-1 relative %s", e.Classes)),
		Attr("x-data", alpineData),
		Attr("@fk-select.window", eventHandler),
		Label(
			Class("label text-sm font-bold flex flex-col items-start gap-1"),
			Text(e.Label),
			Input(
				Type("hidden"), Name(e.Name), Attr(":value", "value"),
				If(e.Required, Required()),
				Iff(e.Attr != nil, func() (out Node) {
					out = Raw("")
					defer func() {
						if r := recover(); r != nil {
							slog.Error("InputForeignKey attr getter panicked", "panic", r, "key", e.Key)
						}
					}()
					n, err := e.Attr(ctx)
					if err != nil {
						slog.Error("InputForeignKey attr getter failed", "error", err, "key", e.Key)
						return out
					}
					if n == nil {
						return out
					}
					v := reflect.ValueOf(n)
					if (v.Kind() == reflect.Pointer || v.Kind() == reflect.Map || v.Kind() == reflect.Slice || v.Kind() == reflect.Interface || v.Kind() == reflect.Func) && v.IsNil() {
						return out
					}
					return n
				}),
			),
			Div(
				Class("flex w-full items-stretch gap-1"),
				Div(
					Class("input input-bordered flex-1 flex items-center cursor-pointer"),
					Attr(":class", "display ? '' : 'opacity-50'"),
					Attr("hx-get", urlStr),
					Attr("hx-target", HTMXTargetBodyModal),
					Attr("hx-swap", HTMXSwapBodyModal),
					Attr("hx-push-url", "false"),
					El("span", Attr("x-text", "display || placeholder")),
				),
				If(
					!e.Required,
					Button(
						Type("button"),
						Class("btn btn-ghost btn-square shrink-0"),
						Attr("@click.stop", "value = ''; display = ''"),
						Attr("x-show", "value"),
						Attr("aria-label", "Clear selection"),
						Render(Icon{Name: "x-mark"}, ctx),
					),
				),
			),
		),
	)
}

// Parse extracts the GORM primary key ID from parameters, queries GORM to verify its database presence, and yields the unit primary key.
func (e InputForeignKey[T]) Parse(v any, ctx context.Context) (any, error) {
	vals, _ := v.([]string)
	if len(vals) == 0 {
		return "", nil
	}
	if strings.TrimSpace(vals[0]) == "" {
		return nil, nil
	}
	i, err := strconv.Atoi(vals[0])
	if err != nil {
		return nil, err
	}
	modelValue := new(T)

	db, err := getters.DBFromContext(ctx)
	if err != nil {
		slog.Error("InputForeignKey: db from context", "error", err)
		return nil, err
	}

	row, err := gorm.G[T](db).Where("ID = ?", i).First(ctx)
	if err != nil {
		slog.Error("Error while fetching data for the specified foreign key", "error", err)
		return nil, err
	}
	*modelValue = row

	return uint(i), nil
}

// GetName returns the HTML form element's name attribute value.
func (e InputForeignKey[T]) GetName() string {
	return e.Name
}
