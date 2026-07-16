package views

import (
	"context"
	"fmt"
	"log/slog"
	"maps"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"sync"

	"github.com/lariv-in/lariv/components"
	"github.com/lariv-in/lariv/getters"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type schemaCacheEntry struct {
	rootSchema   *schema.Schema
	fieldByParam map[string]*schema.Field
}

var (
	schemaCache   = make(map[reflect.Type]*schemaCacheEntry)
	schemaCacheMu sync.RWMutex
)

// LayerList manages paginated database queries and column sorting/filtering operations for collections of type T.
// It parses whitelisted URL query parameters matching model fields, builds dynamic GORM queries,
// loads paged results inside [components.ObjectList], and stores them in the request context under Key.
//
// Automatically applies ILIKE containing checks for string fields, equality matches for other types,
// and resolves ordering clauses depending on the "sort" parameters.
//
// Use Cases:
//   - Displaying search/filter list views (e.g. users directories, transaction histories).
//   - Supporting table widgets with numeric page toggles or filter panels.
//
// Example:
//
//	views.View{
//	    Layers: []views.Layer{
//	        views.LayerList[User]{
//	            Key:      getters.Static("$usersList"),
//	            PageSize: getters.Static(uint(15)),
//	            QueryPatchers: views.QueryPatchers{
//	                views.QueryPatcherPreload[User]("Profile"),
//	            },
//	        },
//	    },
//	}
type LayerList[T any] struct {
	// Key represents the context key string under which the loaded components.ObjectList is stored.
	Key getters.Getter[string]
	// PageSize represents the dynamic Getter returning the number of records per page (defaults to 12).
	PageSize getters.Getter[uint]
	// QueryPatchers represents the slice of query modifiers applied to GORM before retrieving the list.
	QueryPatchers QueryPatchers[T]
}

// Next wraps the downstream HTTP request handlers executing paginated queries.
func (m LayerList[T]) Next(view View, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		db, dberr := getters.DBFromContext(ctx)
		if dberr != nil {
			slog.Error("views: layer list: db from context", "error", dberr)
			ctx = ContextWithErrorsAndValues(ctx, nil, map[string]error{
				"_global": dberr,
			})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		var query gorm.ChainInterface[T] = gorm.G[T](db).Scopes()

		tType := reflect.TypeOf((*T)(nil)).Elem()
		schemaCacheMu.RLock()
		cache, ok := schemaCache[tType]
		schemaCacheMu.RUnlock()
		if !ok {
			schemaCacheMu.Lock()
			cache, ok = schemaCache[tType]
			if !ok {
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
				cache = &schemaCacheEntry{
					rootSchema:   rootSchema,
					fieldByParam: fieldByParam,
				}
				schemaCache[tType] = cache
			}
			schemaCacheMu.Unlock()
		}
		rootSchema := cache.rootSchema
		fieldByParam := cache.fieldByParam
		pageStr := r.URL.Query().Get("page")
		pageNum := uint(1)
		if pageStr != "" {
			if p, err := strconv.ParseUint(pageStr, 10, 32); err == nil && p > 0 {
				pageNum = uint(p)
			}
		}

		var pageSize uint = 12
		var err error
		if m.PageSize != nil {
			pageSize, err = m.PageSize(ctx)
			if err != nil {
				slog.Error("views: layer list: resolve page size", "error", err)
				ctx = ContextWithErrorsAndValues(ctx, nil, map[string]error{
					"_global": fmt.Errorf("failed to resolve page size: %w", err),
				})
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
		}

		// Preserve query params for filters as a $get map, starting with raw URL values.
		queryMap := map[string]any{}
		for param, values := range r.URL.Query() {
			if len(values) > 0 && values[0] != "" {
				queryMap[param] = values[0]
			}
		}

		// If the page has a filter form, parse it to coerce types (e.g., checkboxes to bool)
		// and merge into $get, overriding raw string values where present.
		if page, ok := view.GetPage(); ok {
			if parent, ok := page.(components.ParentInterface); ok {
				if forms := components.FindChildren[components.FormInterface](parent); len(forms) > 0 {
					values, _, perr := forms[0].ParseForm(r)
					if perr != nil {
						slog.Error("views: layer list: parse filter form", "error", perr)
					} else {
						maps.Copy(queryMap, values)
					}
				}
			}
		}

		// Attach $get to the context before any query patching ($request is set by global layer).
		ctx = context.WithValue(ctx, "$get", queryMap)
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
			if f.FieldType.Kind() == reflect.String && f.FieldType.Name() == "string" {
				// Case-insensitive "contains" match for plain strings only.
				// Named string types (e.g. Postgres enums) use equality below.
				query = query.Where(col+" ILIKE ?", "%"+values[0]+"%")
			} else {
				query = query.Where(col+" = ?", values[0])
			}
		}

		query = m.QueryPatchers.Apply(view, r, query)

		// Count assumes sort-driven Joins are BelongsTo/HasOne only (no row duplication);
		// see applyListViewSorts.
		// Qualify the PK column so COUNT is not ambiguous when joins add other tables with "id".
		countCol := "*"
		if rootSchema != nil && rootSchema.Table != "" && rootSchema.PrioritizedPrimaryField != nil {
			if dbn := rootSchema.PrioritizedPrimaryField.DBName; dbn != "" {
				countCol = rootSchema.Table + "." + dbn
			}
		}
		total, err := query.Count(ctx, countCol)
		if err != nil {
			slog.Error("views: layer list: count records", "error", err)
			ctx = ContextWithErrorsAndValues(ctx, nil, map[string]error{
				"_global": fmt.Errorf("failed to count records: %w", err),
			})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		query = query.Limit(int(pageSize)).Offset(int((pageNum - 1) * pageSize))

		results, err := query.Find(ctx)
		if err != nil {
			slog.Error("views: layer list: query records", "error", err)
			ctx = ContextWithErrorsAndValues(ctx, nil, map[string]error{
				"_global": fmt.Errorf("failed to query records: %w", err),
			})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		numPages := ((uint(total) + pageSize - 1) / pageSize)

		objectList := components.ObjectList[T]{
			Items:    results,
			Number:   pageNum,
			NumPages: numPages,
			Total:    uint64(total),
		}

		// Add the object list to the enriched context and render the page.
		key, err := m.Key(ctx)
		if err != nil {
			slog.Error("views: layer detail: resolve context key", "error", err)
			ctx = ContextWithErrorsAndValues(ctx, nil, map[string]error{
				"_global": fmt.Errorf("failed to resolve context key: %w", err),
			})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		ctx = context.WithValue(ctx, key, objectList)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}
