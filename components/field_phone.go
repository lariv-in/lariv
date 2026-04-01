package components

import (
	"context"
	"log/slog"

	"github.com/lariv-in/lago/getters"
	"github.com/nyaruka/phonenumbers"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type FieldPhone struct {
	Page
	Getter  getters.Getter[string]
	Classes string
}

func (e FieldPhone) GetKey() string {
	return e.Key
}

func (e FieldPhone) GetRoles() []string {
	return e.Roles
}

func (e FieldPhone) Build(ctx context.Context) Node {
	if e.Getter == nil {
		return Group{}
	}

	value, err := e.Getter(ctx)
	if err != nil {
		slog.Error("FieldPhone getter failed", "error", err, "key", e.Key)
		return ContainerError{Error: getters.Static(err)}.Build(ctx)
	}

	v, err := phonenumbers.Parse(value, "IN")
	if err != nil {
		return ContainerError{Error: getters.Static(err)}.Build(ctx)
	}

	return Div(Class(e.Classes), Text(phonenumbers.Format(v, phonenumbers.E164)))
}
