package components

import (
	"context"
	"time"

	"github.com/lariv-in/getters"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type FieldDatetime struct {
	Page
	Getter  getters.Getter
	Classes string
}

func (e FieldDatetime) Build(ctx context.Context) Node {
	v, ok := e.Getter(ctx).(time.Time)
	if !ok {
		return Group{}
	}
	timezone := ctx.Value("$tz").(*time.Location)
	if timezone == nil {
		timezone = DefaultTimeZone
	}
	return Div(Class(e.Classes), Text(v.In(timezone).Format("Mon, 02 Jan 2006 15:04:05")))
}
