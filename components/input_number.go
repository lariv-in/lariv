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

type InputNumber[T getters.Number] struct {
	Page
	Label    string
	Name     string
	Getter   getters.Getter[T]
	Required bool
	Classes  string
	// Hidden renders only a hidden input (no label). Parsed value is still T.
	Hidden bool
}

func (e InputNumber[T]) GetKey() string {
	return e.Key
}

func (e InputNumber[T]) GetRoles() []string {
	return e.Roles
}

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
		return Div(Class(wrapClass),
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
	return Div(Class(wrapClass),
		Label(Class("label text-sm font-bold flex flex-col items-start gap-1"),
			Text(e.Label),
			Input(inputAttrs...),
		),
	)
}

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

func (e InputNumber[T]) GetName() string {
	return e.Name
}
