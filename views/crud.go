package views

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/lariv-in/components"
	"github.com/lariv-in/getters"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)


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

					pageStr := r.URL.Query().Get("page")
					pageNum := 1
					if pageStr != "" {
						if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
							pageNum = p
						}
					}
					pageSize := 12

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

					var total int64
					if err := query.Count(&total).Error; err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
						return
					}

					var results []T
					err := query.Limit(pageSize).Offset((pageNum - 1) * pageSize).Find(&results).Error
					if err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
						return
					}

					numPages := int((total + int64(pageSize) - 1) / int64(pageSize))

					objectList := components.ObjectList[T]{
						Items:    results,
						Number:   pageNum,
						NumPages: numPages,
						Total:    total,
					}

					ctx := context.WithValue(r.Context(), key, objectList)
					ctx = context.WithValue(ctx, "$request", r)

					// Preserve query params in context as $get map for filter re-population
					queryMap := map[string]any{}
					for param, values := range r.URL.Query() {
						if len(values) > 0 && values[0] != "" {
							queryMap[param] = values[0]
						}
					}
					ctx = context.WithValue(ctx, "$get", queryMap)

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

					ctx := context.WithValue(r.Context(), key, getters.MapFromStruct(instance))
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
				values, fieldErrors, err := innerView.ParseForm(w, r)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				if HasErrors(fieldErrors) {
					innerView.RenderWithErrors(w, r, fieldErrors, values)
					return
				}

				db := r.Context().Value("$db").(*gorm.DB)

				fmt.Println(values)
				// Create using the map directly with RETURNING to get the generated ID
				err = db.Model(new(T)).Clauses(clause.Returning{Columns: []clause.Column{{Name: "id"}}}).Create(values).Error
				if err != nil {
					ctx := context.WithValue(r.Context(), "$error", map[string]any{"_form": fmt.Errorf("%v", err)})
					innerView.RenderWithErrors(w, r.WithContext(ctx), fieldErrors, values)
					return
				}

				http.Redirect(w, r, fmt.Sprintf(successUrl, values["id"]), http.StatusSeeOther)
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
				values, fieldErrors, err := innerView.ParseForm(w, r)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				if HasErrors(fieldErrors) {
					innerView.RenderWithErrors(w, r, fieldErrors, values)
					return
				}

				idStr := r.PathValue("id")
				id, err := strconv.Atoi(idStr)
				if err != nil {
					http.Error(w, "Invalid ID", http.StatusBadRequest)
					return
				}

				db := r.Context().Value("$db").(*gorm.DB)

				// Update using the map directly, ID already known from path
				err = db.Model(new(T)).Where("id = ?", id).Updates(values).Error
				if err != nil {
					ctx := context.WithValue(r.Context(), "$error", map[string]any{"_form": fmt.Errorf("%v", err)})
					innerView.RenderWithErrors(w, r.WithContext(ctx), fieldErrors, values)
					return
				}

				http.Redirect(w, r, fmt.Sprintf(successUrl, id), http.StatusSeeOther)
			})
		})
	}
}

// --- Singleton Handler ---

// SingletonView loads a singleton record of type T (via FirstOrCreate) into $in context for GET,
// and parses the form + updates the record on POST, then redirects to the URL resolved by successUrl.
func SingletonView[T any](model T, successUrl getters.Getter) func(View) View {
	return func(v View) View {
		// Wrap GET to load singleton into $in context
		oldGet := v.Handlers[http.MethodGet]
		v.Handlers[http.MethodGet] = func(innerView View) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				db := r.Context().Value("$db").(*gorm.DB)
				instance := new(T)
				db.FirstOrCreate(instance)
				ctx := context.WithValue(r.Context(), "$in", getters.MapFromStruct(instance))
				oldGet(innerView).ServeHTTP(w, r.WithContext(ctx))
			})
		}

		// Add POST handler for form save
		return v.WithMethod(http.MethodPost, func(innerView View) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				values, fieldErrors, err := innerView.ParseForm(w, r)
				if err != nil {
					return
				}

				if HasErrors(fieldErrors) {
					innerView.RenderWithErrors(w, r, fieldErrors, values)
					return
				}

				db := r.Context().Value("$db").(*gorm.DB)
				instance := new(T)
				db.FirstOrCreate(instance)

				err = db.Model(instance).Updates(values).Error
				if err != nil {
					ctx := context.WithValue(r.Context(), "$error", map[string]any{"_form": fmt.Errorf("%v", err)})
					innerView.RenderWithErrors(w, r.WithContext(ctx), fieldErrors, values)
					return
				}

				redirectUrl, _ := getters.IfOrGetter(successUrl, r.Context(), "").(string)
				http.Redirect(w, r, redirectUrl, http.StatusSeeOther)
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
