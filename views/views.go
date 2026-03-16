package views

import (
	"context"
	"errors"
	"log/slog"
	"maps"
	"net/http"
	"reflect"

	"github.com/lariv-in/components"
	"github.com/lariv-in/getters"
	"github.com/lariv-in/registry"
	"gorm.io/gorm"
)

type (
	FormPatcher  = func(view *View, r *http.Request, formData map[string]any) map[string]any
	QueryPatcher = func(view *View, r *http.Request, db *gorm.DB) *gorm.DB
)

type Middleware = func(http.Handler) http.Handler

type View struct {
	PageName     string
	Registry     map[string]components.PageInterface
	Handlers     map[string]func(*View) http.Handler
	FormPatcher  FormPatcher
	QueryPatcher QueryPatcher
	Middlewares  registry.Registry[Middleware]
}

func (v *View) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handler, isHandlerPresent := v.Handlers[r.Method]
	if !isHandlerPresent {
		slog.Error("Handler in view was not found")
		http.NotFound(w, r)
		return
	}
	h := handler(v)
	for _, middleware := range *v.Middlewares.AllStable() {
		h = middleware.Value(h)
	}
	h.ServeHTTP(w, r)
}

func (v *View) GetPage() (components.PageInterface, bool) {
	page, isPagePresent := v.Registry[v.PageName]
	return page, isPagePresent
}

func (v *View) RenderPage(w http.ResponseWriter, r *http.Request) {
	page, isPagePresent := v.GetPage()
	if !isPagePresent {
		http.NotFound(w, r)
		return
	}
	ctx := r.Context()

	if shell, ok := page.(components.Shell); ok {
		if isBoosted, _ := ctx.Value("isHtmxBoosted").(bool); isBoosted {
			shell.Body(ctx).Render(w)
			return
		}
	}

	components.Render(page, ctx).Render(w)
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
	if v.FormPatcher != nil {
		values = v.FormPatcher(v, r, values)
	}
	return values, fieldErrors, nil
}

// RenderWithErrors re-renders the view's page with field errors and previously submitted values in context.
func (v *View) RenderWithErrors(w http.ResponseWriter, r *http.Request, fieldErrors map[string]error, values map[string]any) {
	page, _ := v.GetPage()
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
	components.Render(page, ctx).Render(w)
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

func (v *View) WithFormPatcher(formPatcher FormPatcher) *View {
	v.FormPatcher = formPatcher
	return v
}

func (v *View) WithQueryPatcher(queryPatcher QueryPatcher) *View {
	v.QueryPatcher = queryPatcher
	return v
}

func (v *View) WithMiddleware(name string, middleware Middleware) *View {
	v.Middlewares.Register(name, middleware)
	return v
}

func (v *View) WithMiddlewares(middlewares ...registry.Pair[string, Middleware]) *View {
	for _, middleware := range middlewares {
		v.Middlewares.Register(middleware.Key, middleware.Value)
	}
	return v
}

func (v *View) PatchMiddlewares(middlewares ...registry.Pair[string, func(Middleware) Middleware]) *View {
	for _, middleware := range middlewares {
		v.Middlewares.Patch(middleware.Key, middleware.Value)
	}
	return v
}

func (v *View) PatchMiddleware(name string, patcher func(Middleware) Middleware) *View {
	v.Middlewares.Patch(name, patcher)
	return v
}
