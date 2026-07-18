package p_filesystem

import (
	"context"

	"github.com/lariv-in/lariv"
	"github.com/lariv-in/lariv/components"
	"github.com/lariv-in/lariv/getters"
	. "maragu.dev/gomponents"
)

type InputDirectory struct {
	components.Page
	Label       string
	Name        string
	VNode       getters.Getter[VNode]
	Placeholder string
	Required    bool
	Classes     string
}

func (e InputDirectory) GetKey() string {
	return e.Key
}

func (e InputDirectory) GetRoles() []string {
	return e.Roles
}

func (e InputDirectory) Build(ctx context.Context) Node {
	fk := components.InputForeignKey[VNode]{
		Page:        e.Page,
		Label:       e.Label,
		Name:        e.Name,
		Getter:      getters.Getter[VNode](e.VNode),
		Display:     getters.Key[string]("$in.Name"),
		Placeholder: e.Placeholder,
		Url:         lariv.RoutePath("filesystem.SelectRoute", nil),
		Required:    e.Required,
		Classes:     e.Classes,
	}
	return fk.Build(ctx)
}

func (e InputDirectory) Parse(v any, ctx context.Context) (any, error) {
	fk := components.InputForeignKey[VNode]{
		Name: e.Name,
	}
	return fk.Parse(v, ctx)
}

func (e InputDirectory) GetName() string {
	return e.Name
}
