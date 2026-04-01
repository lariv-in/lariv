package components

import (
	"context"
	"log/slog"

	"github.com/lariv-in/lago/getters"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// FieldLink renders a plain text link: href from Href; visible text from Label when set, otherwise the href.
type FieldLink struct {
	Page
	Href    getters.Getter[string]
	Label   getters.Getter[string]
	Classes string
}

func (e FieldLink) GetKey() string {
	return e.Key
}

func (e FieldLink) GetRoles() []string {
	return e.Roles
}

func (e FieldLink) Build(ctx context.Context) Node {
	href := ""
	if e.Href != nil {
		v, err := e.Href(ctx)
		if err != nil {
			slog.Error("FieldLink href getter failed", "error", err, "key", e.Key)
			return ContainerError{Error: getters.Static(err)}.Build(ctx)
		}
		href = v
	}
	label := href
	if e.Label != nil {
		if v, err := e.Label(ctx); err == nil && v != "" {
			label = v
		}
	}
	if href == "" {
		return Div(Class(e.Classes), Text(label))
	}
	return A(Href(href), Class(e.Classes), Text(label))
}
