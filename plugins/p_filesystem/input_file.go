package p_filesystem

import (
	"context"
	"io"

	"github.com/lariv-in/lariv"
	"github.com/lariv-in/lariv/components"
	"github.com/lariv-in/lariv/getters"
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

func (e InputFile) Build(cat components.Catalog, ctx context.Context, w io.Writer) error {
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
	return fk.Build(cat, ctx, w)
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
