package lago

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/registry"
	"github.com/lariv-in/lago/views"
)

var RegistryView registry.Registry[*views.View] = registry.NewRegistry[*views.View]()

type DynamicView struct {
	Key string
}

// #region agent log
func debugLogDynamicView(runID, hypothesisID, location, message string, data map[string]any) {
	f, err := os.OpenFile("/home/sandy/source_repos/lago/.cursor/debug-84938a.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer f.Close()
	_ = json.NewEncoder(f).Encode(map[string]any{
		"sessionId":    "84938a",
		"runId":        runID,
		"hypothesisId": hypothesisID,
		"location":     location,
		"message":      message,
		"data":         data,
		"timestamp":    time.Now().UnixMilli(),
	})
}

// #endregion

func NewDynamicView(key string) DynamicView {
	return DynamicView{Key: key}
}

func (v DynamicView) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	view, viewPresent := RegistryView.Get(v.Key)
	// #region agent log
	debugLogDynamicView("initial", "H1", "lago/registry_views.go:39", "dynamic view lookup", map[string]any{
		"key":         v.Key,
		"viewPresent": viewPresent,
		"path":        r.URL.Path,
		"method":      r.Method,
	})
	// #endregion
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
	}
}
