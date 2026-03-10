package components

import (
	"context"
	"fmt"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type InputForeignKey struct {
	Label       string
	Name        string
	Getter      Getter
	DisplayAttr string
	Placeholder string
	Url         Getter
	Required    bool
	Classes     string
}

func (e InputForeignKey) Build(ctx context.Context) Node {
	value := IfOrGetter(e.Getter, ctx, nil)

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
			if e.DisplayAttr != "" {
				if disp, exists := valueMap[e.DisplayAttr]; exists {
					displayValue = fmt.Sprintf("%v", disp)
				}
			}
		}
	}

	placeholder := e.Placeholder
	if placeholder == "" {
		placeholder = "Select..."
	}

	shown := displayValue
	opacityClass := ""
	if shown == "" {
		shown = placeholder
		opacityClass = " opacity-50"
	}

	return Div(Class(fmt.Sprintf("my-1 %s", e.Classes)),
		Label(Class("label text-sm font-bold"), Text(e.Label)),
		Input(Type("hidden"), Name(e.Name), Value(valuePk),
			If(e.Required, Required())),
		Div(Class("input input-bordered w-full flex items-center cursor-pointer"+opacityClass),
			Text(shown),
		),
	)
}

func (e InputForeignKey) Parse(v string) (any, error) {
	return v, nil
}

func (e InputForeignKey) GetName() string {
	return e.Name
}
