package lago

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	_ "gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func Start(address string, certFile *string, keyFile *string) error {
	db, err := InitDb()
	if err != nil {
		return err
	}

	RegistryMiddleware.Register("core.DbMiddleware", MiddlewareDb(db))
	RegistryMiddleware.Register("core.LoggingMiddlware", MiddlewareLogging)
	RegistryMiddleware.Register("core.HtmxBoostMiddleware", MiddlewareHtmxBoost)

	// Applying all middlewares
	middlewares := RegistryMiddleware.All()
	var router http.Handler = GetRouter()
	for _, middleware := range *middlewares {
		router = middleware(router)
	}

	if certFile != nil && keyFile != nil {
		return http.ListenAndServeTLS(address, *certFile, *keyFile, router)
	}
	if certFile != nil {
		slog.Warn("certFile for tls was not provided")
	}
	if keyFile != nil {
		slog.Warn("keyFile for tls was not provided")
	}
	slog.Warn("Using plain http without tls, ensure this is running in debug or behind a reverse proxy")
	return http.ListenAndServe(address, router)
}

func MiddlewareDb(db *gorm.DB) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), "$db", db)))
		})
	}
}

func MiddlewareLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Use a custom ResponseWriter to capture the status code
		wrapped := &statusWriter{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(wrapped, r)

		slog.Info("http_request",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Int("status", wrapped.status),
			slog.Duration("latency", time.Since(start)),
			slog.String("ip", r.RemoteAddr),
		)
	})
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func MiddlewareHtmxBoost(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		isBoosted := r.Header.Get("HX-Boosted") == "true"
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), "isHtmxBoosted", isBoosted)))
	})
}
