// Package layers contains explanations and code examples for middleware request-handling layers in Lariv.
//
// # Middleware Layers (layers.go)
//
// Middleware layers intercept HTTP requests targeting view handlers to resolve contexts, check security roles, or load DB schemas.
//
// # Existing Built-in Layers
//
//   - views.PathLayer: Extracts route wildcards to the context path map.
//   - views.MethodLayer: Selectively routes HTTP methods to custom sub-handlers.
//   - views.LayerDetail[T]: Loads a single model record T from a route path parameter ID.
//   - views.LayerList[T]: Queries GORM to load, sort, paginate, and list collections of model record T.
//   - views.LayerCreate[T]: Commits new model records T parsing and validating form values.
//   - views.LayerUpdate[T]: Saves updates to loaded database entities T parsing and validating form values.
//   - views.LayerDelete[T]: Removes database entities T.
//   - views.LayerSingleton[T]: Loads a single record T or allocates a blank record.
//   - views.LayerJsonImport[T]: Import CSV/JSON dataset objects.
//   - views.LayerMultiStepForm[T]: Controls wizard wizards page updates.
//   - views.LayerTableToggleColumns[T]: Toggles table column views.
//
// # Creating and Adding a Custom Layer
//
// Custom layers must implement the views.Layer interface:
//
//	package myplugin
//
//	import (
//		"net/http"
//		"log/slog"
//		"github.com/lariv-in/lariv/views"
//	)
//
//	type LoggingLayer struct {
//		Prefix string
//	}
//
//	func (l LoggingLayer) Next(view views.View, next http.Handler) http.Handler {
//		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//			slog.Info(l.Prefix, "url", r.URL.Path)
//			next.ServeHTTP(w, r)
//		})
//	}
//
//	// Registering the custom layer inside views.go:
//	myView := &views.View{
//		Layers: []registry.Pair[string, views.Layer]{
//			{Key: "custom_logger", Value: LoggingLayer{Prefix: "[VIEW-HIT]"}},
//		},
//	}
package layers
