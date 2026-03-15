package components

import (
	"context"
	"fmt"
	"time"

	"github.com/lariv-in/getters"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type FieldDate struct {
	Page
	Getter  getters.Getter
	Classes string
}

func (e FieldDate) Build(ctx context.Context) Node {
	v, ok := e.Getter(ctx).(time.Time)
	if !ok {
		return Group{}
	}
	timezone := ctx.Value("$tz").(*time.Location)
	if timezone == nil {
		timezone = DefaultTimeZone
	}
	return Div(Class(fmt.Sprintf("text-xl font-semibold text-primary %s", e.Classes)), Text(v.In(timezone).Format(time.DateOnly)))
}
