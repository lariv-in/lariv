package components

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/lariv-in/getters"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type InputForeignKey[T any] struct {
	Page
	Label       string
	Name        string
	Getter      getters.Getter[T]
	Display     getters.Getter[string]
	Placeholder string
	Url         getters.Getter[string]
	Required    bool
	Classes     string
}

func (e InputForeignKey[T]) GetKey() string {
	return e.Key
}

func (e InputForeignKey[T]) GetRoles() []string {
	return e.Roles
}

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
				if idVal, exists := valueMap["ID"]; exists {
					valuePk = fmt.Sprintf("%v", idVal)
				} else if idVal, exists := valueMap["id"]; exists {
					valuePk = fmt.Sprintf("%v", idVal)
				}
				if e.Display != nil {
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
		}
	}

	alpineData := fmt.Sprintf("{ value: '%s', display: '%s', placeholder: '%s' }", valuePk, displayValue, placeholder)
	eventHandler := fmt.Sprintf("if ($event.detail.name === '%s') { value = $event.detail.value; display = $event.detail.display; $el.querySelector('.fk-modal-container').innerHTML = ''; }", e.Name)

	return Div(
		Class(fmt.Sprintf("my-1 relative %s", e.Classes)),
		Attr("x-data", alpineData),
		Attr("@fk-select.window", eventHandler),
		Label(Class("label text-sm font-bold"), Text(e.Label)),
		Input(Type("hidden"), Name(e.Name), Attr(":value", "value"),
			If(e.Required, Required())),
		Div(Class("input input-bordered w-full flex items-center cursor-pointer"),
			Attr(":class", "display ? '' : 'opacity-50'"),
			Attr("hx-get", urlStr),
			Attr("hx-target", "next .fk-modal-container"),
			Attr("hx-swap", "innerHTML"),
			Attr("hx-push-url", "false"),
			El("span", Attr("x-text", "display || placeholder")),
		),
		Div(Class("fk-modal-container")),
	)
}

func (e InputForeignKey[T]) Parse(v any, _ context.Context) (any, error) {
	vals, _ := v.([]string)
	if len(vals) == 0 {
		return "", nil
	}
	return vals[0], nil
}

func (e InputForeignKey[T]) GetName() string {
	return e.Name
}
