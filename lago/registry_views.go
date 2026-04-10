package lago

import (
	"context"
	"net/http"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/registry"
	"github.com/lariv-in/lago/views"
	. "maragu.dev/gomponents"
)

var RegistryView *registry.Registry[*views.View] = registry.NewRegistry[*views.View]()

type DynamicView struct {
	Key string
}

func NewDynamicView(key string) DynamicView {
	return DynamicView{Key: key}
}

func (v DynamicView) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	view, viewPresent := RegistryView.Get(v.Key)
	if !viewPresent {
		http.NotFound(w, r)
		return
	}
	view.ServeHTTP(w, r)
}

type redirectPage struct {
	components.Page
}

func (p redirectPage) Build(context.Context) Node {
	return Group{}
}

func (p redirectPage) GetKey() string {
	return p.Page.Key
}

func (p redirectPage) GetRoles() []string {
	return p.Page.Roles
}

func redirectPageLookup(string) (components.PageInterface, bool) {
	return redirectPage{Page: components.Page{Key: "redirect"}}, true
}

type RedirectLayer struct {
	URLGetter getters.Getter[string]
}

func (m RedirectLayer) Next(_ views.View, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		url, err := getters.IfOr(m.URLGetter, r.Context(), "")
		if err != nil || url == "" {
			http.NotFound(w, r)
			return
		}
		views.HtmxRedirect(w, r, url, http.StatusMovedPermanently)
	})
}

func RedirectView(urlGetter getters.Getter[string]) *views.View {
	return &views.View{
		PageName:   "redirect",
		PageLookup: redirectPageLookup,
		Layers: []registry.Pair[string, views.Layer]{
			{Key: "redirect", Value: RedirectLayer{URLGetter: urlGetter}},
		},
	}
}
