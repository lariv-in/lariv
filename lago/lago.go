package lago

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	_ "gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func Start(config LagoConfig) error {
	db, err := InitDB(config)
	if err != nil {
		return err
	}

	RegistryMiddleware.Register("core.DbMiddleware", MiddlewareDb(db))
	RegistryMiddleware.Register("core.LoggingMiddlware", MiddlewareLogging)
	RegistryMiddleware.Register("core.HtmxBoostMiddleware", MiddlewareHtmxBoost)
	RegistryMiddleware.Register("core.EnvironmentMiddleware", MiddlewareEnvironment)

	BuildAllRegistries()

	// Applying all middlewares
	middlewares := RegistryMiddleware.All()
	var router http.Handler = GetRouter()
	for _, middleware := range middlewares {
		router = middleware(router)
	}

	if len(config.CertFile) != 0 && len(config.KeyFile) != 0 {
		return http.ListenAndServeTLS(config.Address, config.CertFile, config.KeyFile, router)
	}

	if len(config.CertFile) != 0 {
		slog.Warn("certFile for tls was not provided")
	}
	if len(config.KeyFile) != 0 {
		slog.Warn("keyFile for tls was not provided")
	}
	slog.Warn("Using plain http without tls, ensure this is running in debug or behind a reverse proxy")
	return http.ListenAndServe(config.Address, router)
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

func MiddlewareEnvironment(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		envMap := map[string]string{}
		if cookie, err := r.Cookie("environment"); err == nil {
			decoded, err := url.QueryUnescape(cookie.Value)
			if err == nil {
				json.Unmarshal([]byte(decoded), &envMap)
			}
		}
		ctx := context.WithValue(r.Context(), "$environment", envMap)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func MiddlewareHtmxBoost(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		isBoosted := r.Header.Get("HX-Boosted") == "true"
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), "isHtmxBoosted", isBoosted)))
	})
}
