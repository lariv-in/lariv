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

// RegistryView represents the global immutable registry tracking view page controllers of type [*views.View].
var RegistryView *registry.ImmutableRegistry[*views.View] = &registry.ImmutableRegistry[*views.View]{}

// DynamicView represents an HTTP handler that lazily resolves and executes a target view from [RegistryView] at request-time.
// This resolves import-cycle constraints between routing endpoints and view logic blocks.
//
// Use Cases:
//   - Delegating routing mappings to plugins without statically importing the views controllers.
//
// Example Definition:
//
//	var HomeHandler = lago.NewDynamicView("core.HomeView")
type DynamicView struct {
	// Key represents the registered identifier of the view controller in RegistryView (e.g., "core.HomeView").
	Key string
}

// NewDynamicView constructs a new DynamicView handler targeting the specified view key.
func NewDynamicView(key string) DynamicView {
	return DynamicView{Key: key}
}

// ServeHTTP satisfies the standard http.Handler interface, executing the resolved view handler.
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

// RedirectLayer represents a view middleware layer that intercepts requests and performs a client redirect.
type RedirectLayer struct {
	// URLGetter represents the dynamic URL target getter resolving to the redirection link.
	URLGetter getters.Getter[string]
}

// Next wraps the down-chain HTTP handler with redirection interceptors.
func (m RedirectLayer) Next(_ views.View, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		url, err := getters.IfOr(m.URLGetter, r.Context(), "")
		if err != nil || url == "" {
			http.NotFound(w, r)
			return
		}
		// See Other: targets are routing decisions, not immutable canonical URLs; avoid caching.
		views.HtmxRedirect(w, r, url, http.StatusSeeOther)
	})
}

// RedirectView returns a specialized [*views.View] instance configured to perform client redirects to the URL resolved by urlGetter.
//
// Use Cases:
//   - Mapping base routes or landing URLs to dashboard landing folders (e.g. mapping "/" to "/dashboard/").
//
// Example:
//
//	var HomeRoute = Route{
//		Path:    "/",
//		Handler: lago.RedirectView(getters.Static("/dashboard/")),
//	}
func RedirectView(urlGetter getters.Getter[string]) *views.View {
	return &views.View{
		PageName:   "redirect",
		PageLookup: redirectPageLookup,
		Layers: []registry.Pair[string, views.Layer]{
			{Key: "redirect", Value: RedirectLayer{URLGetter: urlGetter}},
		},
	}
}
