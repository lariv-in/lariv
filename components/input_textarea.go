package components

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/lariv-in/lago/getters"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type InputTextarea struct {
	Page
	Label    string
	Name     string
	Getter   getters.Getter[string]
	Required bool
	Rows     int
	Classes  string
}

func (e InputTextarea) GetKey() string {
	return e.Key
}

func (e InputTextarea) GetRoles() []string {
	return e.Roles
}

func (e InputTextarea) Build(ctx context.Context) Node {
	rows := e.Rows
	if rows <= 0 {
		rows = 3
	}
	var valueNode Node = Text("")
	if e.Getter != nil {
		value, err := e.Getter(ctx)
		if err != nil {
			slog.Error("InputTextarea getter failed", "error", err, "key", e.Key)
		} else {
			valueNode = Text(value)
		}
	}
	return Div(Class(fmt.Sprintf("my-1 %s", e.Classes)),
		Label(Class("label text-sm font-bold flex flex-col items-start gap-1"),
			Text(e.Label),
			Textarea(Name(e.Name),
				Rows(fmt.Sprintf("%d", rows)),
				valueNode,
				Class(fmt.Sprintf("textarea textarea-bordered w-full %s", e.Classes)),
				If(e.Required, Required())),
		),
	)
}

func (e InputTextarea) Parse(v any, _ context.Context) (any, error) {
	vals, _ := v.([]string)
	if len(vals) == 0 {
		return "", nil
	}
	return vals[0], nil
}

func (e InputTextarea) GetName() string {
	return e.Name
}
