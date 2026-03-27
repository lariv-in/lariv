package views

import (
	"context"
	"net/http"
)

// AttachRequestMiddleware puts the *http.Request in context as "$request" so getters
// can read PathValue on routes that do not set it elsewhere (e.g. outside ListView).
func AttachRequestMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), "$request", r)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// PathMiddleware returns middleware that reads PathValue for each name and stores the
// results in map[string]any under "$path" on the request context.
func PathMiddleware(names ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			m := make(map[string]any, len(names))
			for _, name := range names {
				m[name] = r.PathValue(name)
			}
			ctx := context.WithValue(r.Context(), "$path", m)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
