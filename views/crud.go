package views

import (
	"context"
	"fmt"
	"maps"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/lariv-in/components"
	"github.com/lariv-in/getters"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func ListView[T any](key string) func(*View) *View {
	return func(v *View) *View {
		return v.WithMethod(http.MethodGet, func(innerView *View) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				db := r.Context().Value("$db").(*gorm.DB)
				query := db.Model(new(T))

				// Build a whitelist of allowed filter fields from the GORM schema.
				// Keys are lowercased field / column names, values are the schema fields.
				fieldByParam := map[string]*schema.Field{}
				if stmt := (&gorm.Statement{DB: db}); stmt.Parse(new(T)) == nil && stmt.Schema != nil {
					for _, f := range stmt.Schema.Fields {
						if f.DBName == "" {
							continue
						}
						keyName := strings.ToLower(f.Name)
						keyDB := strings.ToLower(f.DBName)
						fieldByParam[keyName] = f
						fieldByParam[keyDB] = f
					}
				}

				pageStr := r.URL.Query().Get("page")
				pageNum := 1
				if pageStr != "" {
					if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
						pageNum = p
					}
				}
				pageSize := 12

				// Preserve query params for filters as a $get map, starting with raw URL values.
				queryMap := map[string]any{}
				for param, values := range r.URL.Query() {
					if len(values) > 0 && values[0] != "" {
						queryMap[param] = values[0]
					}
				}

				// If the page has a filter form, parse it to coerce types (e.g., checkboxes to bool)
				// and merge into $get, overriding raw string values where present.
				if page, ok := v.GetPage(); ok {
					if parent, ok := page.(components.ParentInterface); ok {
						if forms := components.FindChildren[components.FormInterface](parent); len(forms) > 0 {
							if values, _, err := forms[0].ParseForm(r); err == nil {
								maps.Copy(queryMap, values)
							}
						}
					}
				}

				// Attach $request and $get to the context before any query patching.
				ctx := context.WithValue(r.Context(), "$request", r)
				ctx = context.WithValue(ctx, "$get", queryMap)
				r = r.WithContext(ctx)

				// Apply query param filters using a safe, whitelisted set of columns.
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
					// Look up the field in the allowed map (by struct field name or DB column name).
					f, ok := fieldByParam[strings.ToLower(param)]
					if !ok {
						// Unknown field: ignore rather than constructing raw SQL.
						continue
					}
					col := f.DBName
					if f.FieldType.Kind() == reflect.String {
						// Case-insensitive "contains" match for strings.
						query = query.Where(col+" ILIKE ?", "%"+values[0]+"%")
					} else {
						// Equality match for non-string types.
						query = query.Where(col+" = ?", values[0])
					}
				}

				var total int64
				if err := query.Count(&total).Error; err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				query = query.Limit(pageSize).Offset((pageNum - 1) * pageSize)

				for _, queryPatcher := range v.QueryPatchers {
					query = queryPatcher.Value(v, r, query)
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

				// Add the object list to the enriched context and render the page.
				ctx = context.WithValue(r.Context(), key, objectList)
				innerView.RenderPage(w, r.WithContext(ctx))
			})
		})
	}
}

func DetailView[T any](key string) func(*View) *View {
	return func(v *View) *View {
		return v.WithMethod(http.MethodGet, func(innerView *View) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				idStr := r.PathValue("id")
				id, err := strconv.Atoi(idStr)
				if err != nil {
					http.Error(w, "Invalid ID", http.StatusBadRequest)
					return
				}

				db := r.Context().Value("$db").(*gorm.DB)
				query := db.Model(new(T))
				instance := new(T)
				for _, queryPatcher := range v.QueryPatchers {
					query = queryPatcher.Value(v, r, query)
				}
				err = query.First(instance, id).Error
				if err != nil {
					http.NotFound(w, r)
					return
				}

				// Store the concrete instance under the key so typed GetterKey[T](key)
				// can retrieve it without type errors. Components like Detail[T]
				// will project this into $in as needed.
				ctx := context.WithValue(r.Context(), key, *instance)
				innerView.RenderPage(w, r.WithContext(ctx))
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
				for _, queryPatcher := range innerView.QueryPatchers {
					query = queryPatcher.Value(innerView, r, query)
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
		v.WithMethod(http.MethodGet, func(innerView *View) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				db := r.Context().Value("$db").(*gorm.DB)
				instance := new(T)
				db.FirstOrCreate(instance)
				ctx := context.WithValue(r.Context(), getters.ContextKeyIn, getters.MapFromStruct(instance))
				innerView.RenderPage(w, r.WithContext(ctx))
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
				for _, queryPatcher := range innerView.QueryPatchers {
					query = queryPatcher.Value(innerView, r, query)
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
