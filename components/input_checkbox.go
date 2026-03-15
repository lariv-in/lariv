package components

import (
	"context"
	"fmt"
	"strconv"

	"github.com/lariv-in/getters"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type InputCheckbox struct {
	Page
	Label    string
	Name     string
	Getter   getters.Getter
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
				getters.GetterIf(e.Getter, ctx,
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

func (e InputCheckbox) Parse(v any, _ context.Context) (any, error) {
	vals, _ := v.([]string)
	if len(vals) == 0 {
		return false, nil
	}
	return strconv.ParseBool(vals[0])
}

func (e InputCheckbox) GetKey() string {
	return e.Key
}

func (e InputCheckbox) GetRoles() []string {
	return e.Roles
}

func (e InputCheckbox) GetName() string {
	return e.Name
}
