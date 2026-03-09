package components

import (
	"context"
	"fmt"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type InputPassword struct {
	Label    string
	Name     string
	Getter   Getter
	Required bool
	Classes  string
}

func (e InputPassword) Build(ctx context.Context) Node {
	return Div(Class(fmt.Sprintf("my-1 %s", e.Classes)),
		Label(Class("label text-sm font-bold"), Text(e.Label)),
		Input(Type("password"), Name(e.Name),
			GetterIf(e.Getter, ctx, func(ctx context.Context, value any) Node {
				return Value(fmt.Sprintf("%s", value))
			}),
			Class(fmt.Sprintf("input input-bordered w-full %s", e.Classes)), If(e.Required, Required())),
	)
}

func (e InputPassword) Parse(v string) (any, error) {
	// TODO: Add some password validation here
	return v, nil
}

func (e InputPassword) GetName() string {
	return e.Name
}
