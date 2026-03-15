package lago

import (
	"net/http"

	"github.com/lariv-in/getters"
	"github.com/lariv-in/registry"
)

var RegistryView registry.Registry[http.Handler] = registry.NewRegistry[http.Handler]()

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

// Redirect performs an HTTP redirect that is HTMX-aware.
func Redirect(w http.ResponseWriter, r *http.Request, url string) {
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", url)
		w.WriteHeader(http.StatusOK)
	} else {
		http.Redirect(w, r, url, http.StatusSeeOther)
	}
}

type RedirectView struct {
	RouteKey string
	Args     map[string]getters.Getter
}

func NewRedirectView(routeKey string, args ...map[string]getters.Getter) RedirectView {
	var a map[string]getters.Getter
	if len(args) > 0 {
		a = args[0]
	}
	return RedirectView{RouteKey: routeKey, Args: a}
}

func (v RedirectView) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	getter := GetterRoutePath(v.RouteKey, v.Args)
	url, _ := getters.IfOrGetter(getter, r.Context(), "").(string)
	if url == "" {
		http.NotFound(w, r)
		return
	}
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", url)
		w.WriteHeader(http.StatusOK)
	} else {
		http.Redirect(w, r, url, http.StatusSeeOther)
	}
}
