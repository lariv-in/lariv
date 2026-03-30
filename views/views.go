package views

import (
	"context"
	"errors"
	"log/slog"
	"maps"
	"net/http"
	"reflect"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/registry"
)

type FormPatcher = func(view *View, r *http.Request, formData map[string]any) map[string]any

type Middleware = func(http.Handler) http.Handler

type View struct {
	PageName      string
	PageLookup    func(name string) (components.PageInterface, bool)
	Handlers      map[string]func(*View) http.Handler
	FormPatchers  []registry.Pair[string, FormPatcher]
	QueryPatchers []registry.Pair[string, QueryPatcher]
	// Middlewares are applied in slice order to preserve insertion order.
	Middlewares []registry.Pair[string, Middleware]
	// RenderMiddlewares wrap successful and error renders from CRUD helpers (ListView, DetailView,
	// JsonImport, Create/Update/Singleton POST, etc.). Applied in the same order as Middlewares
	// (earliest registered = outermost).
	RenderMiddlewares []registry.Pair[string, Middleware]
}

type debugResponseWriter struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (w *debugResponseWriter) WriteHeader(statusCode int) {
	w.status = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *debugResponseWriter) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	n, err := w.ResponseWriter.Write(b)
	w.bytes += n
	return n, err
}

func (v *View) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handler, isHandlerPresent := v.Handlers[r.Method]
	if !isHandlerPresent {
		slog.Error("Handler in view was not found")
		http.NotFound(w, r)
		return
	}
	h := handler(v)
	// Apply middlewares in registration order, so that the earliest-registered
	// middleware wraps the handler innermost and the latest wraps outermost.
	for i := len(v.Middlewares) - 1; i >= 0; i-- {
		h = v.Middlewares[i].Value(h)
	}
	h.ServeHTTP(w, r)
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
	v.renderPageResponse(w, r.Context(), page)
}

// renderPageResponse writes HTML for page using the same rules as a successful GET:
// for Shell pages and HTMX-boosted requests, only the shell body is rendered.
func (v *View) renderPageResponse(w http.ResponseWriter, ctx context.Context, page components.PageInterface) {
	dw := &debugResponseWriter{ResponseWriter: w}

	if shell, ok := page.(components.Shell); ok {
		if isBoosted, _ := ctx.Value("isHtmxBoosted").(bool); isBoosted {
			err := shell.Body(ctx).Render(dw)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				slog.Error("Error rendering shell body", "error", err)
				return
			}
			return
		}
	}

	err := components.Render(page, ctx).Render(dw)
	if err != nil {
		panic(err)
	}
}

// ServeRenderPage runs RenderPage after applying RenderMiddlewares. Used by GetPageView and DetailView chaining.
func (v *View) ServeRenderPage(w http.ResponseWriter, r *http.Request) {
	var h http.Handler = http.HandlerFunc(v.RenderPage)
	for i := len(v.RenderMiddlewares) - 1; i >= 0; i-- {
		h = v.RenderMiddlewares[i].Value(h)
	}
	h.ServeHTTP(w, r)
}

// renderWithErrorsWithMiddlewares runs RenderWithErrors wrapped by v.RenderMiddlewares.
func renderWithErrorsWithMiddlewares(v *View, w http.ResponseWriter, r *http.Request, fieldErrors map[string]error, values map[string]any) {
	var h http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		v.RenderWithErrors(w, r, fieldErrors, values)
	})
	for i := len(v.RenderMiddlewares) - 1; i >= 0; i-- {
		h = v.RenderMiddlewares[i].Value(h)
	}
	h.ServeHTTP(w, r)
}

func (v *View) WithMethod(method string, viewHandler func(*View) http.Handler) *View {
	v.Handlers[method] = viewHandler
	return v
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
	for _, formPatcher := range v.FormPatchers {
		values = formPatcher.Value(v, r, values)
	}
	return values, fieldErrors, nil
}

// RenderWithErrors re-renders the view's page with field errors and previously submitted values in context.
func (v *View) RenderWithErrors(w http.ResponseWriter, r *http.Request, fieldErrors map[string]error, values map[string]any) {
	page, isPagePresent := v.GetPage()
	if !isPagePresent {
		http.NotFound(w, r)
		return
	}
	ctx := r.Context()
	errorMap := map[string]any{}
	if existing, ok := ctx.Value(getters.ContextKeyError).(map[string]any); ok {
		maps.Copy(errorMap, existing)
	}
	for name, fieldErr := range fieldErrors {
		if fieldErr != nil {
			errorMap[name] = fieldErr
		}
	}
	ctx = context.WithValue(ctx, getters.ContextKeyError, errorMap)
	inMap := map[string]any{}
	maps.Copy(inMap, values)
	ctx = context.WithValue(ctx, getters.ContextKeyIn, inMap)
	v.renderPageResponse(w, ctx, page)
}

// HasErrors returns true if any error in the map is non-nil.
func (v *View) HasErrors(errs map[string]error) bool {
	for _, err := range errs {
		if err != nil {
			return true
		}
	}
	return false
}

func (v *View) WithFormPatcher(name string, formPatcher FormPatcher) *View {
	v.FormPatchers = append(v.FormPatchers, registry.Pair[string, FormPatcher]{Key: name, Value: formPatcher})
	return v
}

func (v *View) WithQueryPatcher(name string, queryPatcher QueryPatcher) *View {
	v.QueryPatchers = append(v.QueryPatchers, registry.Pair[string, QueryPatcher]{Key: name, Value: queryPatcher})
	return v
}

func (v *View) WithMiddleware(name string, middleware Middleware) *View {
	// Append middleware; keys are labels only and are not required to be unique.
	v.Middlewares = append(v.Middlewares, registry.Pair[string, Middleware]{Key: name, Value: middleware})
	return v
}

// WithRenderMiddleware appends middleware around RenderPage and RenderWithErrors from CRUD view factories.
func (v *View) WithRenderMiddleware(name string, middleware Middleware) *View {
	v.RenderMiddlewares = append(v.RenderMiddlewares, registry.Pair[string, Middleware]{Key: name, Value: middleware})
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
