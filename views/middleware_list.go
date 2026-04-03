package views

import (
	"context"
	"maps"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type MiddlewareList[T any] struct {
	Key           getters.Getter[string]
	PageSize      getters.Getter[uint]
	QueryPatchers QueryPatchers[T]
}

func (m MiddlewareList[T]) Next(view View, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		db := ctx.Value("$db").(*gorm.DB)
		var query gorm.ChainInterface[T] = gorm.G[T](db).Scopes()

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
				panic(err)
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
					if values, _, err := forms[0].ParseForm(r); err == nil {
						maps.Copy(queryMap, values)
					}
				}
			}
		}

		// Attach $get to the context before any query patching ($request is set by global middleware).
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
			if f.FieldType.Kind() == reflect.String {
				// Case-insensitive "contains" match for strings.
				query = query.Where(col+" ILIKE ?", "%"+values[0]+"%")
			} else {
				// Equality match for non-string types.
				query = query.Where(col+" = ?", values[0])
			}
		}

		query = m.QueryPatchers.Apply(view, r, query)

		// Count assumes sort-driven Joins are BelongsTo/HasOne only (no row duplication);
		// see applyListViewSorts.
		total, err := query.Count(ctx, "ID")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		query = query.Limit(int(pageSize)).Offset(int((pageNum - 1) * pageSize))

		results, err := query.Find(ctx)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
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
		ctx = context.WithValue(ctx, m.Key, objectList)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}
