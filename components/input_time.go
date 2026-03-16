package components

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/lariv-in/getters"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type InputTime struct {
	Page
	Label    string
	Name     string
	Getter   getters.Getter[time.Time]
	Required bool
	Classes  string
}

func (e InputTime) Build(ctx context.Context) Node {
	timezone, _ := ctx.Value("$tz").(*time.Location)
	if timezone == nil {
		timezone = DefaultTimeZone
	}
	var valueNode Node = Value("")
	if e.Getter != nil {
		t, err := e.Getter(ctx)
		if err != nil {
			slog.Error("InputTime getter failed", "error", err, "key", e.Key)
			return ContainerError{Error: getters.GetterStatic(err)}.Build(ctx)
		}
		valueNode = Value(t.In(timezone).Format("15:04"))
	}
	return Div(Class(fmt.Sprintf("my-1 %s", e.Classes)),
		Label(Class("label text-sm font-bold"), Text(e.Label)),
		Input(Type("time"), Name(e.Name), valueNode, Class(fmt.Sprintf("input input-bordered w-full %s", e.Classes)), If(e.Required, Required())),
	)
}

func (e InputTime) Parse(v any, ctx context.Context) (any, error) {
	vals, _ := v.([]string)
	if len(vals) == 0 || vals[0] == "" {
		return nil, nil
	}
	timezone, _ := ctx.Value("$tz").(*time.Location)
	if timezone == nil {
		timezone = DefaultTimeZone
	}
	return time.ParseInLocation("15:04", vals[0], timezone)
}

func (e InputTime) GetKey() string {
	return e.Key
}

func (e InputTime) GetRoles() []string {
	return e.Roles
}

func (e InputTime) GetName() string {
	return e.Name
}
