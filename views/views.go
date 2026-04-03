package views

import (
	"errors"
	"log/slog"
	"net/http"
	"reflect"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/registry"
)

type Middleware interface {
	Next(View, http.Handler) http.Handler
}

type View struct {
	PageName    string
	PageLookup  func(name string) (components.PageInterface, bool)
	Middlewares []registry.Pair[string, Middleware]
}

func (v *View) GetHandler() http.Handler {
	var handler http.Handler = http.HandlerFunc(v.RenderPage)
	for i := len(v.Middlewares) - 1; i >= 0; i-- {
		handler = v.Middlewares[i].Value.Next(*v, handler)
	}
	return handler
}

func (v *View) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	v.GetHandler().ServeHTTP(w, r)
}

func (v *View) GetPage() (components.PageInterface, bool) {
	return v.PageLookup(v.PageName)
}

func (v *View) RenderPage(w http.ResponseWriter, r *http.Request) {
	page, isPagePresent := v.GetPage()
	if !isPagePresent {
		http.NotFound(w, r)
		return
	}
	err := components.Render(page, r.Context()).Render(w)
	if err != nil {
		panic(err)
	}
}

// ParseForm finds the first FormComponent in the view's page, parses the request form,
// and returns the values and field errors. Returns true if parsing failed (error already written to w).
func (v *View) ParseForm(w http.ResponseWriter, r *http.Request) (map[string]any, map[string]error, error) {
	page, _ := v.GetPage()
	var parent components.ParentInterface
	// If the page already implements ParentInterface, use it directly.
	if p, ok := page.(components.ParentInterface); ok {
		parent = p
	} else {
		// Many scaffolds (ShellScaffold, ShellTopbarScaffold, etc.) implement
		// ParentInterface only on the pointer type, but are stored as values.
		// For traversal we can safely take the address of the value here.
		val := reflect.ValueOf(page)
		if val.Kind() != reflect.Pointer {
			ptr := val.Addr()
			if p, ok := ptr.Interface().(components.ParentInterface); ok {
				parent = p
			}
		}
	}

	if parent == nil {
		// Log the actual interface assertion issue so it is visible in logs.
		slog.Error("view page does not implement components.ParentInterface",
			"pageType", reflect.TypeOf(page),
			"pageName", v.PageName)
		http.Error(w, "Internal Server Error: No form container found", http.StatusInternalServerError)
		return nil, nil, errors.New("Internal Server Error: No form container found")
	}

	forms := components.FindChildren[components.FormInterface](parent)
	if len(forms) == 0 {
		http.Error(w, "Internal Server Error: No form found", http.StatusInternalServerError)
		return nil, nil, errors.New("Internal Server Error: No form found")
	}
	values, fieldErrors, err := forms[0].ParseForm(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return nil, nil, err
	}
	if fieldErrors == nil {
		fieldErrors = make(map[string]error)
	}
	return values, fieldErrors, nil
}

func (v *View) WithMiddleware(name string, middleware Middleware) *View {
	// Append middleware; keys are labels only and are not required to be unique.
	v.Middlewares = append(v.Middlewares, registry.Pair[string, Middleware]{Key: name, Value: middleware})
	return v
}

// InsertMiddlewareBefore inserts a middleware with the given name immediately
// before the first middleware whose Key matches beforeName. If no such
// middleware exists, it appends it to the end.
func (v *View) InsertMiddlewareBefore(beforeName, name string, middleware Middleware) *View {
	p := registry.Pair[string, Middleware]{Key: name, Value: middleware}
	for i, mw := range v.Middlewares {
		if mw.Key == beforeName {
			v.Middlewares = append(v.Middlewares[:i], append([]registry.Pair[string, Middleware]{p}, v.Middlewares[i:]...)...)
			return v
		}
	}
	// Fallback: behave like WithMiddleware when beforeName is not found.
	return v.WithMiddleware(name, middleware)
}

// InsertMiddlewareAfter inserts a middleware with the given name immediately
// after the first middleware whose Key matches afterName. If no such
// middleware exists, it appends it to the end.
func (v *View) InsertMiddlewareAfter(afterName, name string, middleware Middleware) *View {
	p := registry.Pair[string, Middleware]{Key: name, Value: middleware}
	for i, mw := range v.Middlewares {
		if mw.Key == afterName {
			// Insert after index i.
			pos := i + 1
			if pos >= len(v.Middlewares) {
				v.Middlewares = append(v.Middlewares, p)
			} else {
				v.Middlewares = append(v.Middlewares[:pos], append([]registry.Pair[string, Middleware]{p}, v.Middlewares[pos:]...)...)
			}
			return v
		}
	}
	// Fallback: behave like WithMiddleware when afterName is not found.
	return v.WithMiddleware(name, middleware)
}

func (v *View) WithMiddlewares(middlewares ...registry.Pair[string, Middleware]) *View {
	for _, middleware := range middlewares {
		v.WithMiddleware(middleware.Key, middleware.Value)
	}
	return v
}

func (v *View) PatchMiddlewares(middlewares ...registry.Pair[string, func(Middleware) Middleware]) *View {
	for _, middleware := range middlewares {
		for i, mw := range v.Middlewares {
			if mw.Key == middleware.Key {
				v.Middlewares[i].Value = middleware.Value(mw.Value)
			}
		}
	}
	return v
}

func (v *View) PatchMiddleware(name string, patcher func(Middleware) Middleware) *View {
	for i, mw := range v.Middlewares {
		if mw.Key == name {
			v.Middlewares[i].Value = patcher(mw.Value)
		}
	}
	return v
}
