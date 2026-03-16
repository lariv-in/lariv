package lago

import (
	"net/http"

	"github.com/lariv-in/getters"
	"github.com/lariv-in/registry"
	"github.com/lariv-in/views"
)

var RegistryView registry.Registry[*views.View] = registry.NewRegistry[*views.View]()

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

func NewRedirectView(routeKey string, args ...map[string]getters.Getter[any]) *views.View {
	var a map[string]getters.Getter[any]
	if len(args) > 0 {
		a = args[0]
	}
	redirectHandler := func(_ *views.View) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			getter := GetterRoutePath(routeKey, a)
			url, err := getters.IfOrGetter(getter, r.Context(), "")
			if err != nil || url == "" {
				http.NotFound(w, r)
				return
			}
			if r.Header.Get("HX-Request") == "true" {
				w.Header().Set("HX-Redirect", url)
				w.WriteHeader(http.StatusOK)
			} else {
				http.Redirect(w, r, url, http.StatusSeeOther)
			}
		})
	}
	return &views.View{
		Handlers: map[string]func(*views.View) http.Handler{
			http.MethodGet:  redirectHandler,
			http.MethodPost: redirectHandler,
		},
		Middlewares: registry.NewRegistry[views.Middleware](),
	}
}
