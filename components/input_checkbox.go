package components

import (
	"context"
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
	Hidden   bool
	Attr     getters.Getter[Node]
}

func (e InputCheckbox) Build(ctx context.Context) Node {
	checked := false
	var checkedNode Node = Raw("")
	if e.Getter != nil {
		value, err := e.Getter(ctx)
		if err != nil {
			slog.Error("InputCheckbox getter failed", "error", err, "key", e.Key)
		} else {
			checked = value
			if checked {
				checkedNode = Checked()
			}
		}
	}
	if e.Hidden {
		return Div(
			Class("hidden"),
			Input(
				Type("hidden"),
				Name(e.Name),
				Value(strconv.FormatBool(checked)),
			),
		)
	}
	return Div(
		Class(e.Classes),
		Label(
			Class("label text-sm font-bold cursor-pointer justify-start gap-2 flex flex-row items-center"),
			Input(
				Type("checkbox"),
				If(e.Name != "", Name(e.Name)),
				Value("true"),
				Class("checkbox"),
				If(e.XModel != "", Attr("x-model", e.XModel)),
				checkedNode,
				Iff(e.Attr != nil, func() Node {
					n, err := e.Attr(ctx)
					if err != nil {
						slog.Error("InputCheckbox Attr getter failed", "error", err, "key", e.Key)
						return Raw("")
					}
					return n
				}),
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
