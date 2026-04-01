package components

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/lariv-in/lago/getters"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type InputPassword struct {
	Page
	Label    string
	Name     string
	Getter   getters.Getter[string]
	Required bool
	Classes  string
}

func (e InputPassword) GetKey() string {
	return e.Key
}

func (e InputPassword) GetRoles() []string {
	return e.Roles
}

func (e InputPassword) Build(ctx context.Context) Node {
	valueNode := Value("")
	if e.Getter != nil {
		value, err := e.Getter(ctx)
		if err != nil {
			slog.Error("InputPassword getter failed", "error", err, "key", e.Key)
		} else {
			valueNode = Value(value)
		}
	}
	return Div(Class(fmt.Sprintf("my-1 %s", e.Classes)),
		Label(Class("label text-sm font-bold flex flex-col items-start gap-1"),
			Text(e.Label),
			Input(Type("password"), Name(e.Name), valueNode, Class(fmt.Sprintf("input input-bordered w-full %s", e.Classes)), If(e.Required, Required())),
		),
	)
}

func (e InputPassword) Parse(v any, _ context.Context) (any, error) {
	vals, _ := v.([]string)
	if len(vals) == 0 {
		return "", nil
	}
	// TODO: Add some password validation here
	return vals[0], nil
}

func (e InputPassword) GetName() string {
	return e.Name
}
