package components

import (
	"context"
	"fmt"
	"sort"

	"github.com/lariv-in/lago/getters"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// ClientMatchIf renders one client-side x-if template per match case.
type ClientMatchIf struct {
	Page
	Key      getters.Getter[string]
	Match    getters.Getter[map[string]PageInterface]
	Children []PageInterface
}

func (e ClientMatchIf) Build(ctx context.Context) Node {
	if e.Key == nil {
		return Group{}
	}
	key, err := e.Key(ctx)
	if err != nil {
		return ContainerError{Error: getters.Static(err)}.Build(ctx)
	}
	if e.Match == nil {
		return Group{}
	}
	match, err := e.Match(ctx)
	if err != nil {
		return ContainerError{Error: getters.Static(err)}.Build(ctx)
	}
	keys := make([]string, 0, len(match))
	for k := range match {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	group := Group{}
	for _, k := range keys {
		page := match[k]
		if page == nil {
			continue
		}
		group = append(group,
			El("template",
				Attr("x-if", fmt.Sprintf("%s === %q", key, k)),
				Div(Render(page, ctx)),
			),
		)
	}
	return group
}

func (e ClientMatchIf) GetKey() string {
	return e.Page.Key
}

func (e ClientMatchIf) GetRoles() []string {
	return e.Roles
}

func (e ClientMatchIf) GetChildren() []PageInterface {
	return e.Children
}

func (e *ClientMatchIf) SetChildren(children []PageInterface) {
	e.Children = children
}
