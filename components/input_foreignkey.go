package components

import (
	"context"
	"fmt"

	"github.com/lariv-in/getters"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type InputForeignKey struct {
	Page
	Label       string
	Name        string
	Getter      getters.Getter
	Display     getters.Getter
	Placeholder string
	Url         getters.Getter
	Required    bool
	Classes     string
}

func (e InputForeignKey) GetKey() string {
	return e.Key
}

func (e InputForeignKey) GetRoles() []string {
	return e.Roles
}

func (e InputForeignKey) Build(ctx context.Context) Node {
	value := getters.IfOrGetter(e.Getter, ctx, nil)

	valuePk := ""
	displayValue := ""

	if value != nil {
		valueMap, ok := value.(map[string]any)
		if ok {
			if idVal, exists := valueMap["ID"]; exists {
				valuePk = fmt.Sprintf("%v", idVal)
			} else if idVal, exists := valueMap["id"]; exists {
				valuePk = fmt.Sprintf("%v", idVal)
			}
			if e.Display != nil {
				displayValue = fmt.Sprintf("%v", e.Display(context.WithValue(ctx, "$in", valueMap)))
			}
		}
	}

	placeholder := e.Placeholder
	if placeholder == "" {
		placeholder = "Select..."
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
			Attr("hx-get", fmt.Sprintf("%v", getters.IfOrGetter(e.Url, ctx, ""))),
			Attr("hx-target", "next .fk-modal-container"),
			Attr("hx-swap", "innerHTML"),
			Attr("hx-push-url", "false"),
			El("span", Attr("x-text", "display || placeholder")),
		),
		Div(Class("fk-modal-container")),
	)
}

func (e InputForeignKey) Parse(v any, _ context.Context) (any, error) {
	vals, _ := v.([]string)
	if len(vals) == 0 {
		return "", nil
	}
	return vals[0], nil
}

func (e InputForeignKey) GetName() string {
	return e.Name
}
