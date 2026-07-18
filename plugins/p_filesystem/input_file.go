package p_filesystem

import (
	"context"

	"github.com/lariv-in/lariv"
	"github.com/lariv-in/lariv/components"
	"github.com/lariv-in/lariv/getters"
	. "maragu.dev/gomponents"
)

type InputFile struct {
	components.Page
	Label       string
	Name        string
	VNode       getters.Getter[VNode]
	Placeholder string
	Required    bool
	Classes     string
}

func (e InputFile) GetKey() string {
	return e.Key
}

func (e InputFile) GetRoles() []string {
	return e.Roles
}

func (e InputFile) Build(ctx context.Context) Node {
	fk := components.InputForeignKey[VNode]{
		Page:        e.Page,
		Label:       e.Label,
		Name:        e.Name,
		Getter:      getters.Getter[VNode](e.VNode),
		Display:     getters.Key[string]("$in.Name"),
		Placeholder: e.Placeholder,
		Url:         lariv.RoutePath("filesystem.FileSelectRoute", nil),
		Required:    e.Required,
		Classes:     e.Classes,
	}
	return fk.Build(ctx)
}

func (e InputFile) Parse(v any, ctx context.Context) (any, error) {
	fk := components.InputForeignKey[VNode]{
		Name: e.Name,
	}
	return fk.Parse(v, ctx)
}

func (e InputFile) GetName() string {
	return e.Name
}
