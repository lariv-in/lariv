package components

import (
	"context"
	"fmt"
	"strconv"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type InputCheckbox struct {
	Label    string
	Name     string
	Getter   Getter
	Required bool
	Classes  string
}

func (e InputCheckbox) Build(ctx context.Context) Node {
	return Div(
		Class(fmt.Sprintf("mt-3 %s", e.Classes)),
		Label(
			Class("label cursor-pointer justify-start gap-2"),
			Input(
				Type("checkbox"),
				Name(e.Name),
				Value("true"),
				Class("checkbox"),
				GetterIf(e.Getter, ctx,
					func(ctx context.Context, v any) Node {
						isChecked, isBool := v.(bool)
						if isChecked && isBool {
							return Checked()
						}
						return Raw("")
					},
				),
			),
			Span(Class("label-text"), Text(e.Label)),
		),
	)
}

func (e InputCheckbox) Parse(v string) (any, error) {
	return strconv.ParseBool(v)
}

func (e InputCheckbox) GetName() string {
	return e.Name
}
