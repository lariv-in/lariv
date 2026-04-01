package components

import (
	"context"
	"fmt"
	"log/slog"
	"net/mail"

	"github.com/lariv-in/lago/getters"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type InputEmail struct {
	Page
	Label    string
	Name     string
	Getter   getters.Getter[string]
	Required bool
	Classes  string
}

func (e InputEmail) GetKey() string {
	return e.Key
}

func (e InputEmail) GetRoles() []string {
	return e.Roles
}

func (e InputEmail) Build(ctx context.Context) Node {
	var valueNode Node = Value("")
	if e.Getter != nil {
		value, err := e.Getter(ctx)
		if err != nil {
			slog.Error("InputEmail getter failed", "error", err, "key", e.Key)
		} else {
			valueNode = Value(value)
		}
	}
	return Div(Class(fmt.Sprintf("my-1 %s", e.Classes)),
		Label(Class("label text-sm font-bold flex flex-col items-start gap-1"),
			Text(e.Label),
			Input(Type("email"), Name(e.Name), valueNode, Class(fmt.Sprintf("input input-bordered w-full %s", e.Classes)), If(e.Required, Required())),
		),
	)
}

func (e InputEmail) Parse(v any, _ context.Context) (any, error) {
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
