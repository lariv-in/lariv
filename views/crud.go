package views

import (
	"context"
	"fmt"
	"maps"
	"net/http"
	"strconv"

	"github.com/lariv-in/components"
	"gorm.io/gorm"
)

// parseFormFromView finds the first FormComponent in the view's page tree and parses the request.
func parseFormFromView(v View, r *http.Request) (components.FormComponent, map[string]any, map[string]error, error) {
	page, _ := v.GetPage()
	forms := components.FindChildren[components.FormComponent](page.(components.ParentInterface))
	if len(forms) == 0 {
		return components.FormComponent{}, nil, nil, fmt.Errorf("no form component found in page")
	}
	form := forms[0]
	values, fieldErrors, err := form.ParseForm(r)
	return form, values, fieldErrors, err
}

// renderFormErrors re-renders the page with validation errors and submitted values in context.
func renderFormErrors(v View, w http.ResponseWriter, ctx context.Context, values map[string]any, fieldErrors map[string]error) {
	for name, fieldErr := range fieldErrors {
		if fieldErr != nil {
			ctx = context.WithValue(ctx, "$error."+name, fieldErr)
		}
	}
	for name, value := range values {
		ctx = context.WithValue(ctx, "$in."+name, value)
	}
	page, _ := v.GetPage()
	page.Build(ctx).Render(w)
}

// hasFieldErrors returns true if any field error is non-nil.
func hasFieldErrors(fieldErrors map[string]error) bool {
	for _, err := range fieldErrors {
		if err != nil {
			return true
		}
	}
	return false
}

// --- List View ---

// ListView loads all records from a table into context under the given key.
// Supports query param filtering and sorting.
func ListView(table string, key string) func(View) View {
	return func(v View) View {
		oldHandlers := v.Handlers
		newHandlers := make(map[string]func(View) http.Handler)
		for method, handler := range oldHandlers {
			oldHandler := handler // capture loop variable
			newHandlers[method] = func(innerView View) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					db := r.Context().Value("$db").(*gorm.DB)
					query := db.Table(table)

					// Apply query param filters (simple icontains for strings)
					for param, values := range r.URL.Query() {
						if len(values) == 0 || values[0] == "" {
							continue
						}
						if param == "sort" {
							query = query.Order(values[0])
							continue
						}
						if param == "page" {
							continue
						}
						query = query.Where(fmt.Sprintf("%s LIKE ?", param), "%"+values[0]+"%")
					}

					var results []map[string]any
					err := query.Find(&results).Error
					if err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
						return
					}

					ctx := context.WithValue(r.Context(), key, results)
					oldHandler(innerView).ServeHTTP(w, r.WithContext(ctx))
				})
			}
		}
		v.Handlers = newHandlers
		return v
	}
}

// --- Detail Middleware ---

// DetailView loads a single record by {id} path param into context under the given key.
func DetailView(table string, key string) func(View) View {
	return func(v View) View {
		oldHandlers := v.Handlers
		newHandlers := make(map[string]func(View) http.Handler)
		for method, handler := range oldHandlers {
			oldHandler := handler // capture loop variable
			newHandlers[method] = func(innerView View) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					idStr := r.PathValue("id")
					id, err := strconv.Atoi(idStr)
					if err != nil {
						http.Error(w, "Invalid ID", http.StatusBadRequest)
						return
					}

					db := r.Context().Value("$db").(*gorm.DB)
					result := map[string]any{}
					err = db.Table(table).Where("id = ?", id).Take(&result).Error
					if err != nil {
						http.NotFound(w, r)
						return
					}

					ctx := context.WithValue(r.Context(), key, result)
					oldHandler(innerView).ServeHTTP(w, r.WithContext(ctx))
				})
			}
		}
		v.Handlers = newHandlers
		return v
	}
}

// --- Create Handler ---

// CreateView parses the form, validates, creates a record in the table, and redirects to successUrl.
// successUrl is a format string that receives the new record's ID (e.g. "/users/%v/").
func CreateView(table string, successUrl string) func(View) View {
	return func(v View) View {
		return v.WithMethod(http.MethodPost, func(innerView View) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_, values, fieldErrors, err := parseFormFromView(innerView, r)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				if hasFieldErrors(fieldErrors) {
					renderFormErrors(innerView, w, r.Context(), values, fieldErrors)
					return
				}

				db := r.Context().Value("$db").(*gorm.DB)
				err = db.Table(table).Create(values).Error
				if err != nil {
					ctx := context.WithValue(r.Context(), "$error._form", fmt.Errorf("%v", err))
					renderFormErrors(innerView, w, ctx, values, fieldErrors)
					return
				}

				// Try to get the created ID for redirect
				if id, ok := values["id"]; ok {
					http.Redirect(w, r, fmt.Sprintf(successUrl, id), http.StatusSeeOther)
				} else {
					http.Redirect(w, r, successUrl, http.StatusSeeOther)
				}
			})
		})
	}
}

// --- Update Handler ---

// UpdateView parses the form, validates, updates the record by {id} path param, and redirects.
// successUrl is a format string that receives the record's ID (e.g. "/users/%v/").
func UpdateView(table string, successUrl string) func(View) View {
	return func(v View) View {
		return v.WithMethod(http.MethodPost, func(innerView View) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_, values, fieldErrors, err := parseFormFromView(innerView, r)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				if hasFieldErrors(fieldErrors) {
					renderFormErrors(innerView, w, r.Context(), values, fieldErrors)
					return
				}

				idStr := r.PathValue("id")
				id, err := strconv.Atoi(idStr)
				if err != nil {
					http.Error(w, "Invalid ID", http.StatusBadRequest)
					return
				}

				db := r.Context().Value("$db").(*gorm.DB)
				instance := map[string]any{}
				err = db.Table(table).Where("id = ?", id).Take(&instance).Error
				if err != nil {
					http.NotFound(w, r)
					return
				}

				maps.Copy(instance, values)
				err = db.Table(table).Save(&instance).Error
				if err != nil {
					ctx := context.WithValue(r.Context(), "$error._form", fmt.Errorf("%v", err))
					renderFormErrors(innerView, w, ctx, values, fieldErrors)
					return
				}

				http.Redirect(w, r, fmt.Sprintf(successUrl, id), http.StatusSeeOther)
			})
		})
	}
}

// --- Delete Handler ---

// DeleteView deletes the record by {id} path param and redirects to successUrl.
func DeleteView(table string, successUrl string) func(View) View {
	return func(v View) View {
		return v.WithMethod(http.MethodPost, func(innerView View) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				idStr := r.PathValue("id")
				id, err := strconv.Atoi(idStr)
				if err != nil {
					http.Error(w, "Invalid ID", http.StatusBadRequest)
					return
				}

				db := r.Context().Value("$db").(*gorm.DB)
				err = db.Table(table).Where("id = ?", id).Delete(nil).Error
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				http.Redirect(w, r, successUrl, http.StatusSeeOther)
			})
		})
	}
}
