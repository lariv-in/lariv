package lago

import (
	"log/slog"
	"net/http"

	_ "gorm.io/driver/sqlite"
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
