package components

import (
	"context"
	"fmt"
	"strconv"

	"github.com/lariv-in/getters"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type InputNumber struct {
	Page
	Label    string
	Name     string
	Getter   getters.Getter
	Required bool
	Classes  string
}

func (e InputNumber) Build(ctx context.Context) Node {
	return Div(Class(fmt.Sprintf("my-1 %s", e.Classes)),
		Label(Class("label text-sm font-bold"), Text(e.Label)),
		Input(Type("number"), Name(e.Name),
			getters.GetterIf(e.Getter, ctx, func(ctx context.Context, value any) Node {
				return Value(fmt.Sprintf("%v", value))
			}), Class(fmt.Sprintf("input input-bordered w-full %s", e.Classes)), If(e.Required, Required())),
	)
}

func (e InputNumber) Parse(v any, _ context.Context) (any, error) {
	vals, ok := v.([]string)
	if !ok || len(vals) == 0 || vals[0] == "" {
		return 0, nil
	}
	num, err := strconv.Atoi(vals[0])
	if err != nil {
		return 0, fmt.Errorf("invalid number")
	}
	return num, nil
}

func (e InputNumber) GetName() string {
	return e.Name
}
