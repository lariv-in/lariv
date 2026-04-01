package views

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
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
				var rootSchema *schema.Schema
				fieldByParam := map[string]*schema.Field{}
				if stmt := (&gorm.Statement{DB: db}); stmt.Parse(new(T)) == nil && stmt.Schema != nil {
					rootSchema = stmt.Schema
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

				// Attach $get to the context before any query patching ($request is set by global middleware).
				ctx := context.WithValue(r.Context(), "$get", queryMap)
				r = r.WithContext(ctx)

				// Apply query param filters using a safe, whitelisted set of columns.
				for param, values := range r.URL.Query() {
					if len(values) == 0 {
						continue
					}
					if param == "sort" {
						if rootSchema != nil {
							query = applyListViewSorts(query, rootSchema, values)
						} else {
							var namer schema.Namer = schema.NamingStrategy{}
							if db.Config != nil && db.Config.NamingStrategy != nil {
								namer = db.Config.NamingStrategy
							}
							for _, vv := range values {
								clause := sortQueryValueToOrder(namer, vv)
								if clause != "" {
									query = query.Order(clause)
								}
							}
						}
						continue
					}
					if values[0] == "" {
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

				for _, queryPatcher := range v.QueryPatchers {
					query = queryPatcher.Value(v, r, query)
				}

				// Count assumes sort-driven Joins are BelongsTo/HasOne only (no row duplication);
				// see applyListViewSorts.
				var total int64
				if err := query.Count(&total).Error; err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				query = query.Limit(pageSize).Offset((pageNum - 1) * pageSize)

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
				r = r.WithContext(ctx)
				innerView.ServeRenderPage(w, r)
			})
		})
	}
}

// preloadDetailOrRespond loads T by pathParamKey, applies v.QueryPatchers, stores *instance in context
// under key, and returns the updated request. On failure it writes the HTTP response and returns ok=false.
func preloadDetailOrRespond[T any](v *View, w http.ResponseWriter, r *http.Request, key, pathParamKey string) (*http.Request, bool) {
	idStr := r.PathValue(pathParamKey)
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return nil, false
	}

	db := r.Context().Value("$db").(*gorm.DB)
	query := db.Model(new(T))
	instance := new(T)
	for _, queryPatcher := range v.QueryPatchers {
		query = queryPatcher.Value(v, r, query)
	}
	if err := query.First(instance, id).Error; err != nil {
		http.NotFound(w, r)
		return nil, false
	}

	ctx := context.WithValue(r.Context(), key, *instance)
	return r.WithContext(ctx), true
}

// detailViewWrapNonGETHandler runs preloadDetailOrRespond before the previous handler (e.g. Update POST).
func detailViewWrapNonGETHandler[T any](key, pathParamKey string, old func(*View) http.Handler) func(*View) http.Handler {
	return func(recv *View) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r2, ok := preloadDetailOrRespond[T](recv, w, r, key, pathParamKey)
			if !ok {
				return
			}
			old(recv).ServeHTTP(w, r2)
		})
	}
}

// DetailView loads a record by path {pathParamKey} and stores it in context under key.
// It runs preload on GET (then the previous GET handler or ServeRenderPage) and on every other
// registered method (e.g. POST for UpdateView) so sidebar/detail getters still resolve on
// validation error re-renders.
// When the view already has a GET handler (e.g. from GetPageView or an inner DetailView), GET runs
// preload first, then delegates to that handler so multiple DetailViews can be nested (outer path params first).
// View.RenderMiddlewares apply on the final ServeRenderPage (see WithRenderMiddleware).
func DetailView[T any](key string, pathParamKey string) func(*View) *View {
	return func(v *View) *View {
		prevGET := v.Handlers[http.MethodGet]

		var otherMethods []string
		for m := range v.Handlers {
			if m != http.MethodGet {
				otherMethods = append(otherMethods, m)
			}
		}
		for _, method := range otherMethods {
			old := v.Handlers[method]
			if old == nil {
				continue
			}
			v.Handlers[method] = detailViewWrapNonGETHandler[T](key, pathParamKey, old)
		}

		return v.WithMethod(http.MethodGet, func(innerView *View) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				r2, ok := preloadDetailOrRespond[T](innerView, w, r, key, pathParamKey)
				if !ok {
					return
				}
				r = r2
				if prevGET != nil {
					prevGET(innerView).ServeHTTP(w, r)
					return
				}
				innerView.ServeRenderPage(w, r)
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

func splitAssociationValues(values map[string]any) (map[string]any, map[string]components.AssociationIDs) {
	regularValues := make(map[string]any, len(values))
	associationValues := map[string]components.AssociationIDs{}
	for key, value := range values {
		switch typed := value.(type) {
		case components.AssociationIDs:
			if typed.Field == "" {
				typed.Field = key
			}
			associationValues[key] = typed
		case *components.AssociationIDs:
			if typed == nil {
				continue
			}
			if typed.Field == "" {
				typed.Field = key
			}
			associationValues[key] = *typed
		default:
			regularValues[key] = value
		}
	}
	return regularValues, associationValues
}

func applyAssociationReplacements(db *gorm.DB, record any, associations map[string]components.AssociationIDs) error {
	if len(associations) == 0 {
		return nil
	}

	stmt := &gorm.Statement{DB: db}
	if err := stmt.Parse(record); err != nil {
		return err
	}

	for _, associationValue := range associations {
		relationship, ok := stmt.Schema.Relationships.Relations[associationValue.Field]
		if !ok {
			return fmt.Errorf("unknown association field %q", associationValue.Field)
		}
		if relationship.Type != schema.Many2Many {
			return fmt.Errorf("field %q is not a many-to-many association", associationValue.Field)
		}

		association := db.Model(record).Association(associationValue.Field)
		if association.Error != nil {
			return association.Error
		}

		if len(associationValue.IDs) == 0 {
			if err := association.Clear(); err != nil {
				return err
			}
			continue
		}

		replaceValue, err := buildAssociationReplaceValue(relationship, associationValue.IDs)
		if err != nil {
			return err
		}
		if err := association.Replace(replaceValue); err != nil {
			return err
		}
	}

	return nil
}

func buildAssociationReplaceValue(relationship *schema.Relationship, ids []uint) (any, error) {
	sliceType := relationship.Field.FieldType
	if sliceType.Kind() != reflect.Slice {
		return nil, fmt.Errorf("field %q is not a slice association", relationship.Field.Name)
	}

	elemType := sliceType.Elem()
	elemIsPointer := elemType.Kind() == reflect.Pointer
	baseType := elemType
	if elemIsPointer {
		baseType = elemType.Elem()
	}

	sliceValue := reflect.MakeSlice(sliceType, 0, len(ids))
	for _, id := range ids {
		itemPtr := reflect.New(baseType)
		idField := itemPtr.Elem().FieldByName("ID")
		if !idField.IsValid() || !idField.CanSet() {
			return nil, fmt.Errorf("association %q element type %s does not have a settable ID field", relationship.Field.Name, baseType)
		}
		switch idField.Kind() {
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			idField.SetUint(uint64(id))
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			idField.SetInt(int64(id))
		default:
			return nil, fmt.Errorf("association %q element ID field has unsupported kind %s", relationship.Field.Name, idField.Kind())
		}

		if elemIsPointer {
			sliceValue = reflect.Append(sliceValue, itemPtr)
		} else {
			sliceValue = reflect.Append(sliceValue, itemPtr.Elem())
		}
	}

	return sliceValue.Interface(), nil
}

func modelPrimaryKeyValue(record any) (any, error) {
	value := reflect.ValueOf(record)
	if value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return nil, fmt.Errorf("record is nil")
		}
		value = value.Elem()
	}
	if !value.IsValid() {
		return nil, fmt.Errorf("record is invalid")
	}
	idField := value.FieldByName("ID")
	if !idField.IsValid() {
		return nil, fmt.Errorf("record %T does not have an ID field", record)
	}
	switch idField.Kind() {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return idField.Uint(), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return idField.Int(), nil
	default:
		return nil, fmt.Errorf("record %T has unsupported ID field kind %s", record, idField.Kind())
	}
}

func uploadedJSONFile(values map[string]any, fileField string) (*multipart.FileHeader, error) {
	fileValue, ok := values[fileField]
	if !ok || fileValue == nil {
		return nil, fmt.Errorf("missing %q upload", fileField)
	}

	fileHeader, ok := fileValue.(*multipart.FileHeader)
	if !ok {
		return nil, fmt.Errorf("field %q must be a single file upload", fileField)
	}
	if fileHeader.Filename == "" {
		return nil, fmt.Errorf("field %q did not include a file", fileField)
	}
	if !strings.EqualFold(filepath.Ext(fileHeader.Filename), ".json") {
		return nil, fmt.Errorf("field %q must be a .json file", fileField)
	}
	return fileHeader, nil
}

func decodeJSONArrayFile[T any](fileHeader *multipart.FileHeader) ([]T, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var records []T
	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&records); err != nil {
		return nil, err
	}
	if decoder.More() {
		return nil, fmt.Errorf("json upload must contain exactly one array")
	}
	var trailing json.RawMessage
	if err := decoder.Decode(&trailing); err != nil && err != io.EOF {
		return nil, err
	}
	if len(records) == 0 {
		return []T{}, nil
	}
	return records, nil
}

// JsonImport parses a multipart form, decodes one uploaded .json file into []T,
// creates all rows in a single transaction, and redirects on success.
func JsonImport[T any](fileField string, successURL getters.Getter[string]) func(*View) *View {
	return func(v *View) *View {
		return v.WithMethod(http.MethodPost, func(innerView *View) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				values, fieldErrors, err := innerView.ParseForm(w, r)
				if err != nil {
					renderWithErrorsWithMiddlewares(innerView, w, r, map[string]error{"_form": err}, values)
					return
				}

				if innerView.HasErrors(fieldErrors) {
					renderWithErrorsWithMiddlewares(innerView, w, r, fieldErrors, values)
					return
				}

				fileHeader, err := uploadedJSONFile(values, fileField)
				if err != nil {
					fieldErrors["_form"] = err
					renderWithErrorsWithMiddlewares(innerView, w, r, fieldErrors, values)
					return
				}

				records, err := decodeJSONArrayFile[T](fileHeader)
				if err != nil {
					fieldErrors["_form"] = fmt.Errorf("invalid json import: %w", err)
					renderWithErrorsWithMiddlewares(innerView, w, r, fieldErrors, values)
					return
				}

				db := r.Context().Value("$db").(*gorm.DB)
				if len(records) > 0 {
					if err := db.Transaction(func(tx *gorm.DB) error {
						return tx.Create(&records).Error
					}); err != nil {
						fieldErrors["_form"] = fmt.Errorf("%v", err)
						renderWithErrorsWithMiddlewares(innerView, w, r, fieldErrors, values)
						return
					}
				}

				ctx := context.WithValue(r.Context(), "$count", len(records))
				redirectURL, _ := getters.IfOr(successURL, ctx, "")
				http.Redirect(w, r, redirectURL, http.StatusSeeOther)
			})
		})
	}
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
					renderWithErrorsWithMiddlewares(innerView, w, r, map[string]error{"_form": err}, values)
					return
				}

				if v.HasErrors(fieldErrors) {
					renderWithErrorsWithMiddlewares(innerView, w, r, fieldErrors, values)
					return
				}

				db := r.Context().Value("$db").(*gorm.DB)
				regularValues, associationValues := splitAssociationValues(values)

				record := new(T)
				err = db.Transaction(func(tx *gorm.DB) error {
					if err := PopulateFromMap(record, regularValues); err != nil {
						return err
					}
					if err := tx.Create(record).Error; err != nil {
						return err
					}
					return applyAssociationReplacements(tx, record, associationValues)
				})
				if err != nil {
					fieldErrors["_form"] = fmt.Errorf("%v", err)
					renderWithErrorsWithMiddlewares(innerView, w, r, fieldErrors, values)
					return
				}

				id := uint(reflect.ValueOf(*record).FieldByName("ID").Uint())
				ctx := context.WithValue(r.Context(), "$id", id)
				redirectUrl, _ := getters.IfOr(successURL, ctx, "")
				http.Redirect(w, r, redirectUrl, http.StatusSeeOther)
			})
		})
	}
}

// --- Update Handler ---

// UpdateView parses the form, validates, updates the record by {pathParamKey} path param, and redirects.
// successUrl is a Getter that receives "$id" in context with the record's ID.
func UpdateView[T any](pathParamKey string, successURL getters.Getter[string]) func(*View) *View {
	return func(v *View) *View {
		return v.WithMethod(http.MethodPost, func(innerView *View) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				values, fieldErrors, err := innerView.ParseForm(w, r)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				if innerView.HasErrors(fieldErrors) {
					renderWithErrorsWithMiddlewares(innerView, w, r, fieldErrors, values)
					return
				}

				idStr := r.PathValue(pathParamKey)
				id, err := strconv.ParseUint(idStr, 10, 64)
				if err != nil {
					http.Error(w, "Invalid ID", http.StatusBadRequest)
					return
				}

				db := r.Context().Value("$db").(*gorm.DB)
				regularValues, associationValues := splitAssociationValues(values)
				err = db.Transaction(func(tx *gorm.DB) error {
					query := tx.Model(new(T)).Where("id = ?", id)
					for _, queryPatcher := range innerView.QueryPatchers {
						query = queryPatcher.Value(innerView, r, query)
					}

					record := new(T)
					if err := query.First(record).Error; err != nil {
						return err
					}

					if len(regularValues) > 0 {
						updateQuery := tx.Model(new(T)).Where("id = ?", id)
						for _, queryPatcher := range innerView.QueryPatchers {
							updateQuery = queryPatcher.Value(innerView, r, updateQuery)
						}
						if err := updateQuery.Updates(regularValues).Error; err != nil {
							return err
						}
					}

					return applyAssociationReplacements(tx, record, associationValues)
				})
				if err != nil {
					fieldErrors["_form"] = err
					renderWithErrorsWithMiddlewares(innerView, w, r, fieldErrors, values)
					return
				}

				ctx := context.WithValue(r.Context(), "$id", uint(id))
				redirectUrl, _ := getters.IfOr(successURL, ctx, "")
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
				r = r.WithContext(ctx)
				innerView.ServeRenderPage(w, r)
			})
		})

		return v.WithMethod(http.MethodPost, func(innerView *View) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				values, fieldErrors, err := innerView.ParseForm(w, r)
				if err != nil {
					return
				}

				if innerView.HasErrors(fieldErrors) {
					renderWithErrorsWithMiddlewares(innerView, w, r, fieldErrors, values)
					return
				}

				db := r.Context().Value("$db").(*gorm.DB)
				regularValues, associationValues := splitAssociationValues(values)

				instance := new(T)
				err = db.Transaction(func(tx *gorm.DB) error {
					if err := tx.FirstOrCreate(instance).Error; err != nil {
						return err
					}
					if len(regularValues) > 0 {
						id, err := modelPrimaryKeyValue(instance)
						if err != nil {
							return err
						}
						if err := tx.Model(new(T)).Where("id = ?", id).Updates(regularValues).Error; err != nil {
							return err
						}
					}
					return applyAssociationReplacements(tx, instance, associationValues)
				})
				if err != nil {
					fieldErrors["_form"] = fmt.Errorf("%v", err)
					renderWithErrorsWithMiddlewares(innerView, w, r, fieldErrors, values)
					return
				}

				redirectUrl, _ := getters.IfOr(successURL, r.Context(), "")
				http.Redirect(w, r, redirectUrl, http.StatusSeeOther)
			})
		})
	}
}

// --- Delete Handler ---

// DeleteView deletes the record by {pathParamKey} path param and redirects to successUrl.
func DeleteView[T any](pathParamKey string, successUrl getters.Getter[string]) func(*View) *View {
	return func(v *View) *View {
		return v.WithMethod(http.MethodPost, func(innerView *View) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				idStr := r.PathValue(pathParamKey)
				id, err := strconv.Atoi(idStr)
				if err != nil {
					renderWithErrorsWithMiddlewares(innerView, w, r, map[string]error{"_form": fmt.Errorf("invalid ID")}, nil)
					return
				}

				query := r.Context().Value("$db").(*gorm.DB)
				for _, queryPatcher := range innerView.QueryPatchers {
					query = queryPatcher.Value(innerView, r, query)
				}
				err = query.Delete(new(T), id).Error
				if err != nil {
					renderWithErrorsWithMiddlewares(innerView, w, r, map[string]error{"_form": err}, nil)
					return
				}

				redirectUrl, _ := getters.IfOr(successUrl, r.Context(), "")
				http.Redirect(w, r, redirectUrl, http.StatusSeeOther)
			})
		})
	}
}

// sortQueryValueToOrder maps one URL sort value (e.g. "Name ASC", "User.Name DESC")
// to a GORM Order clause: each path segment is converted with the DB naming strategy
// (snake_case by default), and optional trailing ASC/DESC is preserved.
func sortQueryValueToOrder(namer schema.Namer, raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	parts := strings.Fields(raw)
	if len(parts) == 0 {
		return ""
	}
	dir := ""
	colTokens := parts
	if n := len(parts); n >= 2 {
		last := strings.ToUpper(parts[n-1])
		if last == "ASC" || last == "DESC" {
			dir = " " + last
			colTokens = parts[:n-1]
		}
	}
	if len(colTokens) == 0 {
		return ""
	}
	ident := strings.Join(colTokens, " ")
	colExpr := sortIdentPathToSnakeColumns(namer, ident)
	if colExpr == "" {
		return ""
	}
	return colExpr + dir
}

func sortIdentPathToSnakeColumns(namer schema.Namer, ident string) string {
	segs := strings.Split(ident, ".")
	for i, s := range segs {
		s = strings.TrimSpace(s)
		if s == "" {
			return ""
		}
		segs[i] = namer.ColumnName("", s)
	}
	return strings.Join(segs, ".")
}
