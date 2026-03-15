package components

import (
	"context"
	"fmt"
	"time"

	"github.com/lariv-in/getters"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type InputDatetime struct {
	Page
	Label    string
	Name     string
	Getter   getters.Getter
	Required bool
	Classes  string
}

func (e InputDatetime) GetKey() string {
	return e.Key
}

func (e InputDatetime) GetRoles() []string {
	return e.Roles
}

func (e InputDatetime) Build(ctx context.Context) Node {
	timezone := ctx.Value("$tz").(*time.Location)
	if timezone == nil {
		timezone = DefaultTimeZone
	}
	return Div(Class(fmt.Sprintf("my-1 %s", e.Classes)),
		Label(Class("label text-sm font-bold"), Text(e.Label)),
		Input(Type("datetime-local"), Name(e.Name),
			getters.GetterIf(e.Getter, ctx, func(ctx context.Context, value any) Node {
				if t, ok := value.(time.Time); ok {
					return Value(t.In(timezone).Format("2006-01-02T15:04"))
				}
				return Value(fmt.Sprintf("%s", value))
			}), Class(fmt.Sprintf("input input-bordered w-full %s", e.Classes)), If(e.Required, Required())),
	)
}

func (e InputDatetime) Parse(v any, ctx context.Context) (any, error) {
	vals, _ := v.([]string)
	if len(vals) == 0 || vals[0] == "" {
		return nil, nil
	}
	timezone := ctx.Value("$tz").(*time.Location)
	if timezone == nil {
		timezone = DefaultTimeZone
	}
	return time.ParseInLocation("2006-01-02T15:04", vals[0], timezone)
}

func (e InputDatetime) GetName() string {
	return e.Name
}
