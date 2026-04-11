package components

import (
	"context"

	"github.com/lariv-in/lago/getters"
	. "maragu.dev/gomponents"
)

// GetterPage resolves a page at render time and renders it.
type GetterPage struct {
	Page
	Getter getters.Getter[PageInterface]
}

func (e GetterPage) Build(ctx context.Context) Node {
	if e.Getter == nil {
		return Group{}
	}
	page, err := e.Getter(ctx)
	if err != nil {
		return ContainerError{Error: getters.Static(err)}.Build(ctx)
	}
	if page == nil {
		return Group{}
	}
	return Render(page, ctx)
}

func (e GetterPage) GetKey() string {
	return e.Key
}

func (e GetterPage) GetRoles() []string {
	return e.Roles
}
