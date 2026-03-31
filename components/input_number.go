package components

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/lariv-in/lago/getters"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type InputNumber struct {
	Page
	Label    string
	Name     string
	Getter   getters.Getter[int]
	Required bool
	Classes  string
	// Hidden renders only a hidden input (no label). Parsed value is still int.
	Hidden bool
}

func (e InputNumber) GetKey() string {
	return e.Key
}

func (e InputNumber) GetRoles() []string {
	return e.Roles
}

func (e InputNumber) Build(ctx context.Context) Node {
	var valueNode Node = Value("")
	if e.Getter != nil {
		valueInt, err := e.Getter(ctx)
		if err != nil {
			slog.Error("InputNumber getter failed", "error", err, "key", e.Key)
		} else {
			valueNode = Value(fmt.Sprintf("%d", valueInt))
		}
	}
	wrapClass := fmt.Sprintf("my-1 %s", e.Classes)
	if e.Hidden {
		wrapClass += " hidden"
		return Div(Class(wrapClass),
			Input(Type("hidden"), Name(e.Name), valueNode),
		)
	}
	return Div(Class(wrapClass),
		Label(Class("label text-sm font-bold"), Text(e.Label)),
		Input(Type("number"), Name(e.Name), valueNode, Class(fmt.Sprintf("input input-bordered w-full %s", e.Classes)), If(e.Required, Required())),
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
