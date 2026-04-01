package components

import (
	"context"
	"log/slog"
	"time"

	"github.com/lariv-in/lago/getters"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type FieldTime struct {
	Page
	Getter  getters.Getter[time.Time]
	Classes string
}

func (e FieldTime) GetKey() string {
	return e.Key
}

func (e FieldTime) GetRoles() []string {
	return e.Roles
}

func (e FieldTime) Build(ctx context.Context) Node {
	if e.Getter == nil {
		return Group{}
	}
	v, err := e.Getter(ctx)
	if err != nil {
		slog.Error("FieldTime getter failed", "error", err, "key", e.Key)
		return ContainerError{Error: getters.Static(err)}.Build(ctx)
	}
	timezone, _ := ctx.Value("$tz").(*time.Location)
	if timezone == nil {
		timezone = DefaultTimeZone
	}
	return Div(Class(e.Classes), Text(v.In(timezone).Format(time.TimeOnly)))
}
