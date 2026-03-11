package components

import (
	"context"
	"fmt"

	"github.com/lariv-in/getters"
	"github.com/nyaruka/phonenumbers"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type InputPhone struct {
	Page
	Label    string
	Name     string
	Getter   getters.Getter
	Required bool
	Classes  string
}

func (e InputPhone) Build(ctx context.Context) Node {
	return Div(Class(fmt.Sprintf("my-1 %s", e.Classes)),
		Label(Class("label text-sm font-bold"), Text(e.Label)),
		Input(Type("tel"), Name(e.Name), getters.GetterIf(e.Getter, ctx, func(ctx context.Context, value any) Node {
			return Value(fmt.Sprintf("%s", value))
		}), Class(fmt.Sprintf("input input-bordered w-full %s", e.Classes)), If(e.Required, Required())),
	)
}

func (e InputPhone) Parse(v any) (any, error) {
	vals, _ := v.([]string)
	if len(vals) == 0 {
		return "", nil
	}
	num, err := phonenumbers.Parse(vals[0], "IN")
	if err != nil {
		return nil, err
	}
	return phonenumbers.Format(num, phonenumbers.E164), nil
}

func (e InputPhone) GetName() string {
	return e.Name
}
