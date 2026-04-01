package components

import (
	"context"
	"log/slog"
	"time"

	"github.com/lariv-in/lago/getters"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type FieldDatetime struct {
	Page
	Getter  getters.Getter[time.Time]
	Classes string
}

func (e FieldDatetime) GetKey() string {
	return e.Key
}

func (e FieldDatetime) GetRoles() []string {
	return e.Roles
}

func (e FieldDatetime) Build(ctx context.Context) Node {
	if e.Getter == nil {
		return Group{}
	}
	v, err := e.Getter(ctx)
	if err != nil {
		slog.Error("FieldDatetime getter failed", "error", err, "key", e.Key)
		return ContainerError{Error: getters.Static(err)}.Build(ctx)
	}
	timezone, _ := ctx.Value("$tz").(*time.Location)
	if timezone == nil {
		timezone = DefaultTimeZone
	}
	return Div(Class(e.Classes), Text(v.In(timezone).Format("Mon, 02 Jan 2006 15:04:05")))
}
