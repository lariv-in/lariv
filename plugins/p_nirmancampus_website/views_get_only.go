package p_nirmancampus_website

import (
	"context"
	"net/http"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/registry"
	"github.com/lariv-in/lago/views"
	. "maragu.dev/gomponents"
)

const websiteHandlerPlaceholderPageKey = "nirmancampus_website.handler_placeholder"

// websiteHandlerPlaceholder is a minimal PageInterface for views that only use MethodLayer
// and never render a real page.
type websiteHandlerPlaceholder struct {
	components.Page
}

func (websiteHandlerPlaceholder) Build(context.Context) Node {
	return Group{}
}

func (p websiteHandlerPlaceholder) GetKey() string {
	return p.Page.Key
}

func (p websiteHandlerPlaceholder) GetRoles() []string {
	return p.Page.Roles
}

func websiteHandlerPageLookup(string) (components.PageInterface, bool) {
	return websiteHandlerPlaceholder{Page: components.Page{Key: websiteHandlerPlaceholderPageKey}}, true
}

// websiteGETOnlyView registers a view that handles GET with the given handler and has no other methods.
func websiteGETOnlyView(handler func(*views.View) http.Handler) *views.View {
	return &views.View{
		PageName:   websiteHandlerPlaceholderPageKey,
		PageLookup: websiteHandlerPageLookup,
		Layers: []registry.Pair[string, views.Layer]{
			{Key: "nirmancampus_website.get", Value: views.MethodLayer{
				Method:  http.MethodGet,
				Handler: handler,
			}},
		},
	}
}
