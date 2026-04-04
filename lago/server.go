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

	RegistryLayer.Register("core.AttachRequestLayer", views.AttachRequestLayer{})
	RegistryLayer.Register("core.DbLayer", DBLayer{DB: db})
	if config.Debug {
		RegistryLayer.Register("core.LoggingLayer", LoggingLayer{})
		RegistryLayer.Register("core.CacheDisableLayer", CacheDisableLayer{})
	}
	RegistryLayer.Register("core.HtmxBoostLayer", HtmxBoostLayer{})
	RegistryLayer.Register("core.EnvironmentLayer", EnvironmentLayer{})

	BuildAllRegistries()

	// Applying all layers
	layers := *RegistryLayer.AllStable()
	var router http.Handler = GetRouter(config)
	for _, layer := range layers {
		router = layer.Value.Next(router)
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
