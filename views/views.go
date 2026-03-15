package views

import (
	"context"
	"errors"
	"log/slog"
	"maps"
	"net/http"

	"github.com/lariv-in/components"
	"github.com/lariv-in/getters"
)

type FormPatcher = func(view View, r *http.Request, formData map[string]any) map[string]any

type View struct {
	PageName    string
	Registry    map[string]components.PageInterface
	Handlers    map[string]func(View) http.Handler
	FormPatcher FormPatcher
}

func (v View) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handler, isHandlerPresent := v.Handlers[r.Method]
	if !isHandlerPresent {
		slog.Error("Handler in view was not found")
		http.NotFound(w, r)
		return
	}
	handler(v).ServeHTTP(w, r)
}

func (v View) GetPage() (components.PageInterface, bool) {
	page, isPagePresent := v.Registry[v.PageName]
	return page, isPagePresent
}

func (v View) RenderPage(w http.ResponseWriter, r *http.Request) {
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

func (v View) WithMethod(method string, viewHandler func(View) http.Handler) View {
	v.Handlers[method] = viewHandler
	return v
}

// ParseForm finds the first FormComponent in the view's page, parses the request form,
// and returns the values and field errors. Returns true if parsing failed (error already written to w).
func (v View) ParseForm(w http.ResponseWriter, r *http.Request) (map[string]any, map[string]error, error) {
	page, _ := v.GetPage()
	forms := components.FindChildren[components.FormComponent](page.(components.ParentInterface))
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
func (v View) RenderWithErrors(w http.ResponseWriter, r *http.Request, fieldErrors map[string]error, values map[string]any) {
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
func HasErrors(errs map[string]error) bool {
	for _, err := range errs {
		if err != nil {
			return true
		}
	}
	return false
}

func (v View) WithFormPatcher(formPatcher FormPatcher) View {
	v.FormPatcher = formPatcher
	return v
}
