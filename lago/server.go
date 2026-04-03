package lago

import (
	"log/slog"
	"net"
	"net/http"
	"os"

	"github.com/lariv-in/lago/views"
	_ "gorm.io/driver/sqlite"
)

func StartServer(config LagoConfig) error {
	db, err := InitDB(config)
	if err != nil {
		return err
	}

	RegistryMiddleware.Register("core.AttachRequestMiddleware", views.AttachRequestMiddleware{})
	RegistryMiddleware.Register("core.DbMiddleware", DBMiddleware{DB: db})
	RegistryMiddleware.Register("core.LoggingMiddlware", LoggingMiddleware{})
	RegistryMiddleware.Register("core.HtmxBoostMiddleware", HtmxBoostMiddleware{})
	RegistryMiddleware.Register("core.EnvironmentMiddleware", EnvironmentMiddleware{})
	RegistryMiddleware.Register("core.CacheDisableMiddlware", CacheDisableMiddleware{})

	BuildAllRegistries()

	// Applying all middlewares
	middlewares := *RegistryMiddleware.AllStable()
	var router http.Handler = GetRouter(config)
	for _, middleware := range middlewares {
		router = middleware.Value.Next(router)
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
