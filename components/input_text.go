package components

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/lariv-in/lago/getters"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type InputText struct {
	Page
	Label    string
	Name     string
	Getter   getters.Getter[string]
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
	var valueNode Node = Value("")
	if e.Getter != nil {
		value, err := e.Getter(ctx)
		if err != nil {
			slog.Error("InputText getter failed", "error", err, "key", e.Key)
		} else {
			valueNode = Value(value)
		}
	}
	wrapClass := fmt.Sprintf("my-1 %s", e.Classes)
	if e.Hidden {
		wrapClass += " hidden"
	}
	return Div(Class(wrapClass),
		If(!e.Hidden, Label(Class("label text-sm font-bold"), Text(e.Label))),
		Input(If(!e.Hidden, Type("text")), If(e.Hidden, Type("hidden")), Name(e.Name),
			valueNode, Class(fmt.Sprintf("input input-bordered w-full %s", e.Classes)), If(e.Required, Required())),
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
