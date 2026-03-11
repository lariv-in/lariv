package components

import (
	"context"
	"fmt"
	"net/mail"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type InputEmail struct {
	Page
	Label    string
	Name     string
	Getter   Getter
	Required bool
	Classes  string
}

func (e InputEmail) Build(ctx context.Context) Node {
	return Div(Class(fmt.Sprintf("my-1 %s", e.Classes)),
		Label(Class("label text-sm font-bold"), Text(e.Label)),
		Input(Type("email"), Name(e.Name), GetterIf(e.Getter, ctx, func(ctx context.Context, value any) Node {
			return Value(fmt.Sprintf("%s", value))
		}), Class(fmt.Sprintf("input input-bordered w-full %s", e.Classes)), If(e.Required, Required())),
	)
}

func (e InputEmail) Parse(v any) (any, error) {
	vals, _ := v.([]string)
	if len(vals) == 0 {
		return "", nil
	}
	address, err := mail.ParseAddress(vals[0])
	if err != nil {
		return nil, err
	}
	return address.Address, nil
}

func (e InputEmail) GetName() string {
	return e.Name
}
