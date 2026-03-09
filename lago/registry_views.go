package lago

import (
	"net/http"
)

var RegistryView Registry[http.Handler] = NewRegistry[http.Handler]()

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

type RedirectView struct {
	RouteKey string
}

func NewRedirectView(routeKey string) RedirectView {
	return RedirectView{RouteKey: routeKey}
}

func (v RedirectView) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	route, routePresent := RegistryRoute.Get(v.RouteKey)
	if !routePresent {
		http.NotFound(w, r)
		return
	}
	http.Redirect(w, r, route.Path, http.StatusTemporaryRedirect)
}
