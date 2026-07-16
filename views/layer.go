package views

import (
	"context"
	"net/http"
	"time"
)

// Layer defines a view-specific middleware execution layer.
// It wraps an HTTP handler chain to inspect, modify, or intercept HTTP requests.
//
// Use Cases:
//   - Parsing view-specific parameters (e.g. path values) or checking custom access policies.
//   - Injecting telemetry hooks or headers targeting individual views.
//
// Example:
//
//	type HeaderInjectorLayer struct{}
//
//	func (l HeaderInjectorLayer) Next(view views.View, next http.Handler) http.Handler {
//		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//			w.Header().Set("X-Custom-View-Header", "Lariv")
//			next.ServeHTTP(w, r)
//		})
//	}
type Layer interface {
	// Next wraps the next handler in the view request execution chain.
	Next(View, http.Handler) http.Handler
}

// GlobalLayer defines a global HTTP middleware layer that wraps the base server route multiplexer.
type GlobalLayer interface {
	// Next wraps the global HTTP handler.
	Next(http.Handler) http.Handler
}

func requestQueryMap(r *http.Request) map[string]any {
	queryMap := map[string]any{}
	for param, values := range r.URL.Query() {
		if len(values) > 0 && values[0] != "" {
			queryMap[param] = values[0]
		}
	}
	return queryMap
}

// AttachRequestLayer injects the *http.Request context as "$request", the raw query parameter map as "$get",
// and the current Unix microseconds timestamp as "$timestamp".
// It is registered globally inside lariv.StartServer.
type AttachRequestLayer struct{}

// Next executes the request wrapping functionality injecting standard context values.
func (AttachRequestLayer) Next(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), "$request", r)
		ctx = context.WithValue(ctx, "$get", requestQueryMap(r))
		ctx = context.WithValue(ctx, "$timestamp", time.Now().UnixMicro())
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// PathLayer extracts URL path parameter variables (PathValue) and stores them in a map[string]any
// under the "$path" context key.
type PathLayer struct {
	// Names represents the slice of URL path parameter keys to extract.
	Names []string
}

// Next executes the extraction layer, mapping parameters and invoking the next handler.
func (m PathLayer) Next(_ View, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		values := make(map[string]any, len(m.Names))
		for _, name := range m.Names {
			values[name] = r.PathValue(name)
		}
		ctx := context.WithValue(r.Context(), "$path", values)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// MethodLayer routes incoming requests matching a specific HTTP Method string
// directly to a custom sub-handler function, bypassing subsequent middleware chains.
type MethodLayer struct {
	// Method represents the target HTTP verb (e.g., "GET", "POST").
	Method string
	// Handler represents the routing sub-handler builder function.
	Handler func(*View) http.Handler
}

// Next inspects the request method, routing to the Method handler if matched.
func (m MethodLayer) Next(view View, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == m.Method {
			m.Handler(&view).ServeHTTP(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}
