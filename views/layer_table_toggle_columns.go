package views

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/lariv-in/lariv/components"
	"github.com/lariv-in/lariv/getters"
)

// LayerTableToggleColumns parses a URL query parameter carrying a list of active columns,
// builds a boolean map of visible columns, and stores it in the request context under ContextKey.
// Downstream data tables fetch this map (e.g., using [components.GetterEnabledColumnsFromContext]) to dynamically show or hide fields.
//
// Use Cases:
//   - Supporting customizable table layouts where users can check/uncheck columns via toolbar toggles.
//
// Example:
//
//	views.View{
//	    Layers: []views.Layer{
//	        views.LayerTableToggleColumns{
//	            QueryParam: getters.Static("cols"),
//	            ContextKey: getters.Static("visibleColumnsMap"),
//	        },
//	        views.LayerList[User]{
//	            Key: getters.Static("$users"),
//	        },
//	    },
//	}
type LayerTableToggleColumns struct {
	// QueryParam represents the Getter resolving to the URL query parameter key (e.g. "cols").
	QueryParam getters.Getter[string]
	// ContextKey represents the Getter resolving to the request context key under which the visibility map is saved.
	ContextKey getters.Getter[string]
}

// Next wraps the downstream HTTP request handlers executing table columns parsing.
func (m LayerTableToggleColumns) Next(_ View, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		paramName, err := m.QueryParam(ctx)
		if err != nil {
			slog.Error("views: LayerTableToggleColumns: query param name", "error", err)
			next.ServeHTTP(w, r)
			return
		}
		keyName, err := m.ContextKey(ctx)
		if err != nil {
			slog.Error("views: LayerTableToggleColumns: context key", "error", err)
			next.ServeHTTP(w, r)
			return
		}
		q := r.URL.Query()
		if _, ok := q[paramName]; !ok {
			next.ServeHTTP(w, r)
			return
		}
		raw := ""
		if vals := q[paramName]; len(vals) > 0 {
			raw = vals[0]
		}
		parsed := components.ParseEnabledTableColumnsParam(raw)
		ctx = context.WithValue(ctx, keyName, parsed)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
