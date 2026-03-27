package lago

import (
	"log/slog"
	"net"
	"net/http"
	"os"

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
	var router http.Handler = GetRouter(config)
	for _, middleware := range middlewares {
		router = middleware(router)
	}

	slog.Warn("Using plain http without tls, ensure this is running in debug or behind a reverse proxy")
	if config.UDS != "" {
		if err := os.Remove(config.UDS); err != nil && !os.IsNotExist(err) {
			return err
		}
		ln, err := net.Listen("unix", config.UDS)
		if err != nil {
			return err
		}
		err = os.Chmod(config.UDS, 0o777)
		if err != nil {
			return err
		}
		defer ln.Close()
		return http.Serve(ln, router)
	}
	return http.ListenAndServe(config.Address, router)
}
