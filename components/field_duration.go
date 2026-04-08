package components

import (
	"context"
	"log/slog"
	"time"

	"github.com/lariv-in/lago/getters"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type FieldDuration struct {
	Page
	Getter  getters.Getter[*time.Duration]
	Classes string
}

func (e FieldDuration) GetKey() string {
	return e.Key
}

func (e FieldDuration) GetRoles() []string {
	return e.Roles
}

func (e FieldDuration) Build(ctx context.Context) (out Node) {
	out = Group{}
	if e.Getter == nil {
		return out
	}
	defer func() {
		if r := recover(); r != nil {
			slog.Error("FieldDuration getter panicked", "panic", r, "key", e.Key)
			out = Div(Class(e.Classes), Text(""))
		}
	}()
	v, err := e.Getter(ctx)
	if err != nil {
		slog.Error("FieldDuration getter failed", "error", err, "key", e.Key)
		return ContainerError{Error: getters.Static(err)}.Build(ctx)
	}
	if v == nil {
		return Div(Class(e.Classes), Text(""))
	}
	return Div(Class(e.Classes), Text(v.String()))
}
