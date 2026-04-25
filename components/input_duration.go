package components

import (
	"context"
	"fmt"
	"log/slog"
	"reflect"
	"strings"
	"time"

	"github.com/lariv-in/lago/getters"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// InputDuration renders a text input for Go duration strings (e.g. "30s", "5m", "1h30m").
// Parse returns *time.Duration; empty input returns nil.
type InputDuration struct {
	Page
	Label    string
	Name     string
	Getter   getters.Getter[*time.Duration]
	Required bool
	Classes  string
	Hidden   bool
	Attr     getters.Getter[Node]
}

func (e InputDuration) GetKey() string {
	return e.Key
}

func (e InputDuration) GetRoles() []string {
	return e.Roles
}

func (e InputDuration) Build(ctx context.Context) Node {
	var valueNode Node = Value("")
	if e.Getter != nil {
		d, err := e.Getter(ctx)
		if err != nil {
			slog.Error("InputDuration getter failed", "error", err, "key", e.Key)
		} else if d != nil {
			valueNode = Value(d.String())
		}
	}

	wrapClass := fmt.Sprintf("my-1 %s", e.Classes)
	if e.Hidden {
		wrapClass += " hidden"
	}

	return Div(Class(wrapClass),
		Label(Class("label text-sm font-bold flex flex-col items-start gap-1"),
			If(!e.Hidden, Text(e.Label)),
			Input(If(!e.Hidden, Type("text")), If(e.Hidden, Type("hidden")), Name(e.Name),
				valueNode,
				Class(fmt.Sprintf("input input-bordered w-full %s", e.Classes)),
				If(e.Required, Required()),
				Iff(e.Attr != nil, func() (out Node) {
					out = Raw("")
					defer func() {
						if r := recover(); r != nil {
							slog.Error("InputDuration attr getter panicked", "panic", r, "key", e.Key)
						}
					}()
					n, err := e.Attr(ctx)
					if err != nil {
						slog.Error("InputDuration attr getter failed", "error", err, "key", e.Key)
						return out
					}
					if n == nil {
						return out
					}
					v := reflect.ValueOf(n)
					if (v.Kind() == reflect.Pointer || v.Kind() == reflect.Map || v.Kind() == reflect.Slice || v.Kind() == reflect.Interface || v.Kind() == reflect.Func) && v.IsNil() {
						return out
					}
					return n
				}),
			),
		),
	)
}

func (e InputDuration) Parse(v any, _ context.Context) (any, error) {
	vals, _ := v.([]string)
	if len(vals) == 0 {
		return (*time.Duration)(nil), nil
	}
	raw := strings.TrimSpace(vals[0])
	if raw == "" {
		return (*time.Duration)(nil), nil
	}
	d, err := time.ParseDuration(raw)
	if err != nil {
		return nil, fmt.Errorf("invalid duration")
	}
	return new(d), nil
}

func (e InputDuration) GetName() string {
	return e.Name
}
