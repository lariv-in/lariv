package components

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/lariv-in/lago/getters"
	"github.com/nyaruka/phonenumbers"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type InputPhone struct {
	Page
	Label    string
	Name     string
	Getter   getters.Getter[string]
	Required bool
	Classes  string
}

func (e InputPhone) GetKey() string {
	return e.Key
}

func (e InputPhone) GetRoles() []string {
	return e.Roles
}

func (e InputPhone) Build(ctx context.Context) Node {
	displayValue := ""
	if e.Getter != nil {
		value, err := e.Getter(ctx)
		if err != nil {
			slog.Error("InputPhone getter failed", "error", err, "key", e.Key)
		} else {
			if value != "" {
				parsed, err := phonenumbers.Parse(value, "IN")
				if err == nil {
					displayValue = phonenumbers.Format(parsed, phonenumbers.E164)
				} else {
					displayValue = value
				}
			}
		}
	}
	return Div(Class(fmt.Sprintf("my-1 %s", e.Classes)),
		Label(Class("label text-sm font-bold flex flex-col items-start gap-1"),
			Text(e.Label),
			Input(Type("tel"), Name(e.Name), Value(displayValue), Class(fmt.Sprintf("input input-bordered w-full %s", e.Classes)), If(e.Required, Required())),
		),
	)
}

func (e InputPhone) Parse(v any, _ context.Context) (any, error) {
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
