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

type InputCheckbox struct {
	Page
	Label    string
	Name     string
	Getter   getters.Getter[bool]
	XModel   string
	Required bool
	Classes  string
}

func (e InputCheckbox) Build(ctx context.Context) Node {
	var checkedNode Node = Raw("")
	if e.Getter != nil {
		checked, err := e.Getter(ctx)
		if err != nil {
			slog.Error("InputCheckbox getter failed", "error", err, "key", e.Key)
		} else {
			if checked {
				checkedNode = Checked()
			}
		}
	}
	return Div(
		Class(fmt.Sprintf("mt-3 %s", e.Classes)),
		Label(
			Class("label cursor-pointer justify-start gap-2"),
			Input(
				Type("checkbox"),
				Name(e.Name),
				Value("true"),
				Class("checkbox"),
				If(e.XModel != "", Attr("x-model", e.XModel)),
				checkedNode,
			),
			Span(Class("label-text"), Text(e.Label)),
		),
	)
}

func (e InputCheckbox) Parse(v any, _ context.Context) (any, error) {
	vals, _ := v.([]string)
	if len(vals) == 0 {
		return false, nil
	}
	return strconv.ParseBool(vals[0])
}

func (e InputCheckbox) GetKey() string {
	return e.Key
}

func (e InputCheckbox) GetRoles() []string {
	return e.Roles
}

func (e InputCheckbox) GetName() string {
	return e.Name
}
