package components

import (
	"context"
	"time"

	"github.com/lariv-in/getters"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type FieldTime struct {
	Page
	Getter  getters.Getter
	Classes string
}

func (e FieldTime) Build(ctx context.Context) Node {
	v, ok := e.Getter(ctx).(time.Time)
	if !ok {
		return Group{}
	}
	timezone := ctx.Value("$tz").(*time.Location)
	if timezone == nil {
		timezone = DefaultTimeZone
	}
	return Div(Class(e.Classes), Text(v.In(timezone).Format(time.TimeOnly)))
}
