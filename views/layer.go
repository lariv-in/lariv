package views

import (
	"context"
	"net/http"
	"time"
)

type Layer interface {
	Next(View, http.Handler) http.Handler
}

type GlobalLayer interface {
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

// AttachRequestLayer puts the *http.Request in context as "$request", the raw
// query params as "$get", and the current time as int64 Unix microseconds as
// "$timestamp". It is registered on the global HTTP stack in lago.StartServer
// (core.AttachRequestLayer); do not attach "$request" manually in view handlers.
type AttachRequestLayer struct{}

func (AttachRequestLayer) Next(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), "$request", r)
		ctx = context.WithValue(ctx, "$get", requestQueryMap(r))
		ctx = context.WithValue(ctx, "$timestamp", time.Now().UnixMicro())
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// PathLayer returns layer that reads PathValue for each name and stores the
// results in map[string]any under "$path" on the request context.
type PathLayer struct {
	Names []string
}

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

type MethodLayer struct {
	Method  string
	Handler func(*View) http.Handler
}

func (m MethodLayer) Next(view View, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == m.Method {
			m.Handler(&view).ServeHTTP(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}
