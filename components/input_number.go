package components

import (
	"context"
	"fmt"
	"log/slog"
	"reflect"
	"strconv"
	"strings"

	"github.com/lariv-in/lago/getters"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// InputNumber represents a generic numerical value input form field component.
// It supports all numeric types constraints (int, uint, float) via type parameter T.
// It automatically configures input modes (e.g. `decimal` vs `numeric`) and applies standard limits (e.g. `min="0"` for unsigned uints).
//
// Use Cases:
//   - Inputting integer sequence fields, age inputs, currency decimal amounts, or item product quantities.
//
// Example:
//
//	&components.InputNumber[uint]{
//	    Label:  "Quantity",
//	    Name:   "quantity",
//	    Getter: getters.Key[uint]("$in.Quantity"),
//	}
type InputNumber[T getters.Number] struct {
	// Page embeds common component properties like Key and Roles.
	Page
	// Label represents the header label text displayed above the number input.
	Label string
	// Name represents the HTML form parameter name attribute.
	Name string
	// Getter is the dynamic function retrieving the default/current numeric value of type T.
	Getter getters.Getter[T]
	// Required is a boolean indicating if this form number is a mandatory input.
	Required bool
	// Classes represents additional CSS classes applied to the output HTML wrapper.
	// (Discouraged: Use layout containers or theme styling instead of custom styling overrides).
	Classes string
	// Hidden specifies if this number field is rendered as a hidden input element.
	Hidden bool
}

// GetKey returns the unique key identifier for this InputNumber component.
func (e InputNumber[T]) GetKey() string {
	return e.Key
}

// GetRoles returns the authorized roles required to view this InputNumber.
func (e InputNumber[T]) GetRoles() []string {
	return e.Roles
}

// Build compiles the InputNumber component into a Div wrapping a numeric Input.
func (e InputNumber[T]) Build(ctx context.Context) Node {
	var valueNode Node = Value("")
	if e.Getter != nil {
		valueNumber, err := e.Getter(ctx)
		if err != nil {
			slog.Error("InputNumber getter failed", "error", err, "key", e.Key)
		} else {
			valueNode = Value(fmt.Sprintf("%v", valueNumber))
		}
	}
	wrapClass := fmt.Sprintf("my-1 %s", e.Classes)
	if e.Hidden {
		wrapClass += " hidden"
		return Div(
			Class(wrapClass),
			Input(Type("hidden"), Name(e.Name), valueNode),
		)
	}
	var zero T
	kind := reflect.TypeOf(zero).Kind()
	inputAttrs := []Node{
		Type("number"),
		Name(e.Name),
		valueNode,
		Class(fmt.Sprintf("input input-bordered w-full %s", e.Classes)),
		If(e.Required, Required()),
	}
	switch kind {
	case reflect.Float32, reflect.Float64:
		inputAttrs = append(inputAttrs, Attr("inputmode", "decimal"))
	default:
		inputAttrs = append(inputAttrs, Attr("inputmode", "numeric"))
	}
	switch kind {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		inputAttrs = append(inputAttrs, Attr("min", "0"))
	}
	return Div(
		Class(wrapClass),
		Label(
			Class("label text-sm font-bold flex flex-col items-start gap-1"),
			Text(e.Label),
			Input(inputAttrs...),
		),
	)
}

// Parse extracts and parses numerical values from input parameters using reflection.
func (e InputNumber[T]) Parse(v any, _ context.Context) (any, error) {
	var zero T
	vals, ok := v.([]string)
	if !ok || len(vals) == 0 || vals[0] == "" {
		return zero, nil
	}
	raw := strings.TrimSpace(vals[0])
	targetType := reflect.TypeOf(zero)
	value := reflect.New(targetType).Elem()

	switch targetType.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		num, err := strconv.ParseInt(raw, 10, targetType.Bits())
		if err != nil {
			return zero, fmt.Errorf("invalid number")
		}
		value.SetInt(num)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		num, err := strconv.ParseUint(raw, 10, targetType.Bits())
		if err != nil {
			return zero, fmt.Errorf("invalid number")
		}
		value.SetUint(num)
	case reflect.Float32, reflect.Float64:
		num, err := strconv.ParseFloat(raw, targetType.Bits())
		if err != nil {
			return zero, fmt.Errorf("invalid number")
		}
		value.SetFloat(num)
	default:
		return zero, fmt.Errorf("unsupported number type")
	}
	return value.Interface().(T), nil
}

// GetName returns the HTML form element's name attribute value.
func (e InputNumber[T]) GetName() string {
	return e.Name
}
