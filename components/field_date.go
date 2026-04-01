package components

import (
	"context"
	"log/slog"
	"time"

	"github.com/lariv-in/lago/getters"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type FieldDate struct {
	Page
	Getter  getters.Getter[time.Time]
	Classes string
}

func (e FieldDate) GetKey() string {
	return e.Key
}

func (e FieldDate) GetRoles() []string {
	return e.Roles
}

func (e FieldDate) Build(ctx context.Context) Node {
	if e.Getter == nil {
		return Group{}
	}
	v, err := e.Getter(ctx)
	if err != nil {
		slog.Error("FieldDate getter failed", "error", err, "key", e.Key)
		return ContainerError{Error: getters.Static(err)}.Build(ctx)
	}
	timezone, _ := ctx.Value("$tz").(*time.Location)
	if timezone == nil {
		timezone = DefaultTimeZone
	}
	return Time(Class(e.Classes), DateTime(v.In(timezone).Format(time.DateOnly)), Text(v.In(timezone).Format(time.DateOnly)))
}
