package forms

import (
	"context"
	"net/http"
	"strconv"

	"github.com/lariv-in/lago/components"
	"gorm.io/gorm"
)

// AttachFormFieldsObjectListContext adds FormFieldsObjectListContextKey for the fields DataTable
// when context already holds the loaded Form under key "form" (e.g. forms.DetailView GET).
func AttachFormFieldsObjectListContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		formAny := r.Context().Value("form")
		form, ok := formAny.(Form)
		if !ok {
			next.ServeHTTP(w, r)
			return
		}
		items := form.FormFields
		n := len(items)
		ol := components.ObjectList[FormField]{
			Items:    items,
			Number:   1,
			NumPages: 1,
			Total:    int64(n),
		}
		ctx := context.WithValue(r.Context(), FormFieldsObjectListContextKey, ol)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// formIDFromPathContext returns the parent form id from $path.form_id (see views.PathMiddleware).
func formIDFromPathContext(ctx context.Context) (uint, bool) {
	m, ok := ctx.Value("$path").(map[string]any)
	if !ok || m == nil {
		return 0, false
	}
	raw, ok := m["form_id"]
	if !ok || raw == nil {
		return 0, false
	}
	s, ok := raw.(string)
	if !ok || s == "" {
		return 0, false
	}
	u, err := strconv.ParseUint(s, 10, 64)
	if err != nil || u == 0 {
		return 0, false
	}
	return uint(u), true
}

// AttachFormForParentFieldsPath loads Form into context as "form" for field create and submissions list
// (routes without DetailView[Form] loading the parent form). Requires PathMiddleware so $path is populated.
func AttachFormForParentFieldsPath(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		formID, ok := formIDFromPathContext(r.Context())
		if !ok {
			next.ServeHTTP(w, r)
			return
		}
		db, ok := r.Context().Value("$db").(*gorm.DB)
		if !ok || db == nil {
			next.ServeHTTP(w, r)
			return
		}
		form, err := gorm.G[Form](db).Where("id = ?", formID).First(r.Context())
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}
		ctx := context.WithValue(r.Context(), "form", form)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
