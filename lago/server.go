package lago

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/lariv-in/views"
	_ "gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func StartServer(config LagoConfig) error {
	db, err := InitDB(config)
	if err != nil {
		return err
	}

	RegistryMiddleware.Register("core.DbMiddleware", MiddlewareDB(db))
	RegistryMiddleware.Register("core.LoggingMiddlware", MiddlewareLogging)
	RegistryMiddleware.Register("core.HtmxBoostMiddleware", MiddlewareHtmxBoost)
	RegistryMiddleware.Register("core.EnvironmentMiddleware", MiddlewareEnvironment)
	RegistryMiddleware.Register("core.CacheDisableMiddlware", MiddlewareCacheDisable)

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

func MiddlewareDB(db *gorm.DB) views.Middleware {
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
		cookie, err := r.Cookie("environment")
		if err != nil {
			slog.Error("MiddlewareEnvironment: failed to get environment cookie", "error", err)
		} else {
			decoded, err := url.QueryUnescape(cookie.Value)
			if err == nil {
				if err := json.Unmarshal([]byte(decoded), &envMap); err != nil {
					slog.Error("MiddlewareEnvironment: failed to unmarshal environment cookie", "error", err, "cookie", cookie.Value)
				}
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

func MiddlewareCacheDisable(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, proxy-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")

		w.Header().Del("ETag")
		w.Header().Del("Last-Modified")
		next.ServeHTTP(w, r)
	})
}
