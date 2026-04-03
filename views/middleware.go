package views

import (
	"context"
	"net/http"
	"time"
)

type Middleware interface {
	Next(View, http.Handler) http.Handler
}

type GlobalMiddleware interface {
	Next(http.Handler) http.Handler
}

// AttachRequestMiddleware puts the *http.Request in context as "$request" and the
// current time as int64 Unix microseconds as "$timestamp". It is registered on the
// global HTTP stack in lago.StartServer (core.AttachRequestMiddleware); do not attach
// "$request" manually in view handlers.
type AttachRequestMiddleware struct{}

func (AttachRequestMiddleware) Next(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), "$request", r)
		ctx = context.WithValue(ctx, "$timestamp", time.Now().UnixMicro())
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// PathMiddleware returns middleware that reads PathValue for each name and stores the
// results in map[string]any under "$path" on the request context.
type PathMiddleware struct {
	Names []string
}

func (m PathMiddleware) Next(_ View, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		values := make(map[string]any, len(m.Names))
		for _, name := range m.Names {
			values[name] = r.PathValue(name)
		}
		ctx := context.WithValue(r.Context(), "$path", values)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

type MethodMiddleware struct {
	Method  string
	Handler func(*View) http.Handler
}

func (m MethodMiddleware) Next(view View, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == m.Method {
			m.Handler(&view).ServeHTTP(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}
