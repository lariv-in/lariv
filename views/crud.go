package views

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"

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

// getID extracts the ID field from a struct instance using reflection.
func getID(instance any) any {
	v := reflect.ValueOf(instance)
	if v.Kind() == reflect.Pointer {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil
	}
	// Walk embedded structs (e.g. gorm.Model) to find ID
	return findField(v, "ID")
}

func findField(v reflect.Value, name string) any {
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.Name == name {
			return v.Field(i).Interface()
		}
		if field.Anonymous && v.Field(i).Kind() == reflect.Struct {
			if result := findField(v.Field(i), name); result != nil {
				return result
			}
		}
	}
	return nil
}

// applyValues copies form values (snake_case keys) into exported struct fields.
// It walks embedded structs and converts string values to the target field type.
func applyValues(instance any, values map[string]any) {
	v := reflect.ValueOf(instance)
	if v.Kind() == reflect.Pointer {
		v = v.Elem()
	}
	setFields(v, values)
}

func setFields(v reflect.Value, values map[string]any) {
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.Anonymous && v.Field(i).Kind() == reflect.Struct {
			setFields(v.Field(i), values)
			continue
		}
		if field.PkgPath != "" || !v.Field(i).CanSet() {
			continue
		}
		key := toSnakeCase(field.Name)
		val, ok := values[key]
		if !ok {
			continue
		}
		fieldVal := v.Field(i)
		rv := reflect.ValueOf(val)
		if rv.Type().AssignableTo(fieldVal.Type()) {
			fieldVal.Set(rv)
		} else if rv.Type().ConvertibleTo(fieldVal.Type()) {
			fieldVal.Set(rv.Convert(fieldVal.Type()))
		} else if rv.Kind() == reflect.String {
			setFieldFromString(fieldVal, val.(string))
		}
	}
}

func setFieldFromString(field reflect.Value, s string) {
	switch field.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if n, err := strconv.ParseInt(s, 10, 64); err == nil {
			field.SetInt(n)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if n, err := strconv.ParseUint(s, 10, 64); err == nil {
			field.SetUint(n)
		}
	case reflect.Float32, reflect.Float64:
		if f, err := strconv.ParseFloat(s, 64); err == nil {
			field.SetFloat(f)
		}
	case reflect.Bool:
		if b, err := strconv.ParseBool(s); err == nil {
			field.SetBool(b)
		}
	case reflect.String:
		field.SetString(s)
	}
}

func toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			if i > 0 {
				result.WriteByte('_')
			}
			result.WriteRune(r + ('a' - 'A'))
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// --- List View ---

// ListView loads all records for model T into context under the given key.
// Supports query param filtering and sorting.
func ListView[T any](model T, key string) func(View) View {
	return func(v View) View {
		oldHandlers := v.Handlers
		newHandlers := make(map[string]func(View) http.Handler)
		for method, handler := range oldHandlers {
			oldHandler := handler // capture loop variable
			newHandlers[method] = func(innerView View) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					db := r.Context().Value("$db").(*gorm.DB)
					query := db.Model(new(T))

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

					var results []T
					err := query.Find(&results).Error
					if err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
						return
					}

					// Convert structs to maps for template compatibility
					mapped := make([]map[string]any, len(results))
					for i, item := range results {
						mapped[i] = components.MapFromStruct(item)
					}

					ctx := context.WithValue(r.Context(), key, mapped)
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
func DetailView[T any](model T, key string) func(View) View {
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
					instance := new(T)
					err = db.First(instance, id).Error
					if err != nil {
						http.NotFound(w, r)
						return
					}

					ctx := context.WithValue(r.Context(), key, components.MapFromStruct(instance))
					oldHandler(innerView).ServeHTTP(w, r.WithContext(ctx))
				})
			}
		}
		v.Handlers = newHandlers
		return v
	}
}

// --- Create Handler ---

// CreateView parses the form, validates, creates a record of type T, and redirects to successUrl.
// successUrl is a format string that receives the new record's ID (e.g. "/users/%v/").
func CreateView[T any](model T, successUrl string) func(View) View {
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
				instance := new(T)
				applyValues(instance, values)
				result := db.Create(instance)
				if result.Error != nil {
					ctx := context.WithValue(r.Context(), "$error._form", fmt.Errorf("%v", result.Error))
					renderFormErrors(innerView, w, ctx, values, fieldErrors)
					return
				}

				http.Redirect(w, r, fmt.Sprintf(successUrl, getID(instance)), http.StatusSeeOther)
			})
		})
	}
}

// --- Update Handler ---

// UpdateView parses the form, validates, updates the record by {id} path param, and redirects.
// successUrl is a format string that receives the record's ID (e.g. "/users/%v/").
func UpdateView[T any](model T, successUrl string) func(View) View {
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
				instance := new(T)
				err = db.First(instance, id).Error
				if err != nil {
					http.NotFound(w, r)
					return
				}

				applyValues(instance, values)
				err = db.Save(instance).Error
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
func DeleteView[T any](model T, successUrl string) func(View) View {
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
				err = db.Delete(new(T), id).Error
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				http.Redirect(w, r, successUrl, http.StatusSeeOther)
			})
		})
	}
}
