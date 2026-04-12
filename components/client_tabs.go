package components

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/lariv-in/lago/getters"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// ClientTabs renders responsive client-side tabs with a mobile top ribbon and desktop left ribbon.
type ClientTabs struct {
	Page
	Tabs        map[string]getters.Getter[PageInterface]
	Default     getters.Getter[string]
	StateKey    string
	Attr        getters.Getter[Node]
	RibbonAttr  getters.Getter[Node]
	ContentAttr getters.Getter[Node]
}

func (e ClientTabs) Build(ctx context.Context) Node {
	if len(e.Tabs) == 0 {
		return Group{}
	}

	keys := make([]string, 0, len(e.Tabs))
	match := make(map[string]PageInterface, len(e.Tabs))
	for key, pageGetter := range e.Tabs {
		if pageGetter == nil {
			continue
		}
		page, err := pageGetter(ctx)
		if err != nil {
			return ContainerError{Error: getters.Static(err)}.Build(ctx)
		}
		if page == nil {
			continue
		}
		keys = append(keys, key)
		match[key] = page
	}
	if len(keys) == 0 {
		return Group{}
	}
	sort.Strings(keys)

	stateKey := e.StateKey
	if stateKey == "" {
		stateKey = "tab"
	}

	defaultTab := keys[0]
	if e.Default != nil {
		if selected, err := e.Default(ctx); err != nil {
			return ContainerError{Error: getters.Static(err)}.Build(ctx)
		} else if _, ok := match[selected]; ok {
			defaultTab = selected
		}
	}
	xData, err := json.Marshal(map[string]string{stateKey: defaultTab})
	if err != nil {
		return ContainerError{Error: getters.Static(err)}.Build(ctx)
	}

	ribbon := Group{}
	for _, key := range keys {
		ribbon = append(ribbon, Button(
			Type("button"),
			Class("btn flex-1 md:flex-none md:w-full justify-center md:justify-start"),
			Attr("@click", fmt.Sprintf("%s = %q", stateKey, key)),
			Attr(":class", fmt.Sprintf("%s === %q ? 'btn-primary' : 'btn-ghost'", stateKey, key)),
			Text(key),
		))
	}

	return Div(
		Class("flex flex-col gap-4 md:flex-row md:items-start"),
		Attr("x-data", string(xData)),
		Iff(e.Attr != nil, func() Node {
			n, err := e.Attr(ctx)
			if err != nil {
				return ContainerError{Error: getters.Static(err)}.Build(ctx)
			}
			if n == nil {
				return Group{}
			}
			return n
		}),
		Div(
			Class("flex w-full flex-row gap-1 rounded-box border border-base-300 bg-base-100 p-1 md:sticky md:top-2 md:w-56 md:flex-col"),
			Iff(e.RibbonAttr != nil, func() Node {
				n, err := e.RibbonAttr(ctx)
				if err != nil {
					return ContainerError{Error: getters.Static(err)}.Build(ctx)
				}
				if n == nil {
					return Group{}
				}
				return n
			}),
			ribbon,
		),
		Div(
			Class("min-w-0 flex-1"),
			Iff(e.ContentAttr != nil, func() Node {
				n, err := e.ContentAttr(ctx)
				if err != nil {
					return ContainerError{Error: getters.Static(err)}.Build(ctx)
				}
				if n == nil {
					return Group{}
				}
				return n
			}),
			Render(ClientMatchIf{
				Key:   getters.Static(stateKey),
				Match: getters.Static(match),
			}, ctx),
		),
	)
}

func (e ClientTabs) GetKey() string {
	return e.Key
}

func (e ClientTabs) GetRoles() []string {
	return e.Roles
}
