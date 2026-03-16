package views

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"strconv"

	"github.com/go-viper/mapstructure/v2"
	"github.com/lariv-in/components"
	"github.com/lariv-in/getters"
	"gorm.io/gorm"
)

const (
	listViewMiddlewareName      = "views.crud.list"
	detailViewMiddlewareName    = "views.crud.detail"
	singletonViewMiddlewareName = "views.crud.singleton"
)

// --- List View ---

// ListView loads all records for model T into context under the given key.
// Supports query param filtering and sorting.
func ListView[T any](key string) func(*View) *View {
	return func(v *View) *View {
		return v.WithMiddleware(listViewMiddlewareName, func(next http.Handler) http.Handler {
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
					query = query.Where(param, values[0])
				}

				var total int64
				if err := query.Count(&total).Error; err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				query = query.Limit(pageSize).Offset((pageNum - 1) * pageSize)

				if v.QueryPatcher != nil {
					query = v.QueryPatcher(v, r, query)
				}

				var results []T
				err := query.Find(&results).Error
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

				next.ServeHTTP(w, r.WithContext(ctx))
			})
		})
	}
}

// --- Detail Middleware ---

// DetailView loads a single record by {id} path param into context under the given key.
func DetailView[T any](key string) func(*View) *View {
	return func(v *View) *View {
		return v.WithMiddleware(detailViewMiddlewareName, func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				idStr := r.PathValue("id")
				id, err := strconv.Atoi(idStr)
				if err != nil {
					http.Error(w, "Invalid ID", http.StatusBadRequest)
					return
				}

				query := r.Context().Value("$db").(*gorm.DB)
				instance := new(T)
				if v.QueryPatcher != nil {
					query = v.QueryPatcher(v, r, query)
				}
				err = query.First(instance, id).Error
				if err != nil {
					http.NotFound(w, r)
					return
				}

				ctx := context.WithValue(r.Context(), key, getters.MapFromStruct(instance))
				next.ServeHTTP(w, r.WithContext(ctx))
			})
		})
	}
}

func PopulateFromMap[T any](v *T, values map[string]any) error {
	decodeConfig := mapstructure.DecoderConfig{Result: v, Deep: true, Squash: true}
	decoder, err := mapstructure.NewDecoder(&decodeConfig)
	if err != nil {
		return err
	}
	return decoder.Decode(values)
}

// --- Create Handler ---

// CreateView parses the form, validates, creates a record of type T, and redirects to successUrl.
// successUrl is a Getter that receives "$id" in context with the new record's ID.
func CreateView[T any](successURL getters.Getter[string]) func(*View) *View {
	return func(v *View) *View {
		return v.WithMethod(http.MethodPost, func(innerView *View) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				values, fieldErrors, err := innerView.ParseForm(w, r)
				if err != nil {
					innerView.RenderWithErrors(w, r, map[string]error{"_form": err}, values)
					return
				}

				if v.HasErrors(fieldErrors) {
					innerView.RenderWithErrors(w, r, fieldErrors, values)
					return
				}

				db := r.Context().Value("$db").(*gorm.DB)

				record := new(T)
				if err = PopulateFromMap(record, values); err != nil {
					fieldErrors["_form"] = fmt.Errorf("%v", err)
					innerView.RenderWithErrors(w, r, fieldErrors, values)
					return
				}
				err = db.Create(record).Error
				if err != nil {
					fieldErrors["_form"] = fmt.Errorf("%v", err)
					innerView.RenderWithErrors(w, r, fieldErrors, values)
					return
				}

				id := reflect.ValueOf(*record).FieldByName("ID").Uint()
				ctx := context.WithValue(r.Context(), "$id", fmt.Sprintf("%d", id))
				redirectUrl, _ := getters.IfOrGetter(successURL, ctx, "")
				http.Redirect(w, r, redirectUrl, http.StatusSeeOther)
			})
		})
	}
}

// --- Update Handler ---

// UpdateView parses the form, validates, updates the record by {id} path param, and redirects.
// successUrl is a Getter that receives "$id" in context with the record's ID.
func UpdateView[T any](successURL getters.Getter[string]) func(*View) *View {
	return func(v *View) *View {
		return v.WithMethod(http.MethodPost, func(innerView *View) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				values, fieldErrors, err := innerView.ParseForm(w, r)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				if innerView.HasErrors(fieldErrors) {
					innerView.RenderWithErrors(w, r, fieldErrors, values)
					return
				}

				idStr := r.PathValue("id")
				id, err := strconv.Atoi(idStr)
				if err != nil {
					http.Error(w, "Invalid ID", http.StatusBadRequest)
					return
				}

				query := r.Context().Value("$db").(*gorm.DB).Model(new(T)).Where("id = ?", id)
				if innerView.QueryPatcher != nil {
					query = innerView.QueryPatcher(innerView, r, query)
				}

				// Update using the map directly, ID already known from path
				err = query.Updates(values).Error
				if err != nil {
					fieldErrors["_form"] = err
					innerView.RenderWithErrors(w, r, fieldErrors, values)
					return
				}

				ctx := context.WithValue(r.Context(), "$id", fmt.Sprintf("%d", id))
				redirectUrl, _ := getters.IfOrGetter(successURL, ctx, "")
				http.Redirect(w, r, redirectUrl, http.StatusSeeOther)
			})
		})
	}
}

// --- Singleton Handler ---

// SingletonView loads a singleton record of type T (via FirstOrCreate) into $in context for GET,
// and parses the form + updates the record on POST, then redirects to the URL resolved by successUrl.
func SingletonView[T any](successURL getters.Getter[string]) func(*View) *View {
	return func(v *View) *View {
		v.WithMiddleware(singletonViewMiddlewareName, func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					next.ServeHTTP(w, r)
					return
				}

				db := r.Context().Value("$db").(*gorm.DB)
				instance := new(T)
				db.FirstOrCreate(instance)
				ctx := context.WithValue(r.Context(), getters.ContextKeyIn, getters.MapFromStruct(instance))
				next.ServeHTTP(w, r.WithContext(ctx))
			})
		})

		return v.WithMethod(http.MethodPost, func(innerView *View) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				values, fieldErrors, err := innerView.ParseForm(w, r)
				if err != nil {
					return
				}

				if innerView.HasErrors(fieldErrors) {
					innerView.RenderWithErrors(w, r, fieldErrors, values)
					return
				}

				db := r.Context().Value("$db").(*gorm.DB)
				instance := new(T)
				db.FirstOrCreate(instance)

				err = db.Model(instance).Updates(values).Error
				if err != nil {
					fieldErrors["_form"] = fmt.Errorf("%v", err)
					innerView.RenderWithErrors(w, r, fieldErrors, values)
					return
				}

				redirectUrl, _ := getters.IfOrGetter(successURL, r.Context(), "")
				http.Redirect(w, r, redirectUrl, http.StatusSeeOther)
			})
		})
	}
}

// --- Delete Handler ---

// DeleteView deletes the record by {id} path param and redirects to successUrl.
func DeleteView[T any](successUrl getters.Getter[string]) func(*View) *View {
	return func(v *View) *View {
		return v.WithMethod(http.MethodPost, func(innerView *View) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				idStr := r.PathValue("id")
				id, err := strconv.Atoi(idStr)
				if err != nil {
					http.Error(w, "Invalid ID", http.StatusBadRequest)
					return
				}

				query := r.Context().Value("$db").(*gorm.DB)
				if innerView.QueryPatcher != nil {
					query = innerView.QueryPatcher(innerView, r, query)
				}
				err = query.Delete(new(T), id).Error
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				redirectUrl, _ := getters.IfOrGetter(successUrl, r.Context(), "")
				http.Redirect(w, r, redirectUrl, http.StatusSeeOther)
			})
		})
	}
}
