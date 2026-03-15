package components

import (
	"context"
	"fmt"

	"github.com/lariv-in/getters"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type InputText struct {
	Page
	Label    string
	Name     string
	Getter   getters.Getter
	Required bool
	Classes  string
	Hidden   bool
}

func (e InputText) GetKey() string {
	return e.Key
}

func (e InputText) GetRoles() []string {
	return e.Roles
}

func (e InputText) Build(ctx context.Context) Node {
	return Div(Class(fmt.Sprintf("my-1 %s", e.Classes)),
		Label(Class("label text-sm font-bold"), Text(e.Label)),
		Input(If(!e.Hidden, Type("text")), If(e.Hidden, Type("hidden")), Name(e.Name),
			getters.GetterIf(e.Getter, ctx, func(ctx context.Context, value any) Node {
				return Value(fmt.Sprintf("%s", value))
			}), Class(fmt.Sprintf("input input-bordered w-full %s", e.Classes)), If(e.Required, Required())),
	)
}

func (e InputText) Parse(v any, _ context.Context) (any, error) {
	vals, _ := v.([]string)
	if len(vals) == 0 {
		return "", nil
	}
	return vals[0], nil
}

func (e InputText) GetName() string {
	return e.Name
}
