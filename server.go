package lariv

import (
	"log/slog"
	"net"
	"net/http"
	"os"

	_ "gorm.io/driver/sqlite"
)

// StartServer compiles global middleware layers, configures CORS/CrossOrigin protection, and listens for HTTP requests.
// It supports binding to standard TCP socket addresses (config.Address) or Unix Domain Sockets (config.UDS).
//
// Use Cases:
//   - Initializing the primary web server listener block during application startup.
//
// Example:
//
//	if err := lariv.StartServer(config); err != nil {
//		log.Fatal(err)
//	}
func StartServer(config LarivConfig) error {
	// Applying all layers
	layers := *RegistryLayer.AllStable()
	var router http.Handler = GetRouter(config)
	for _, layer := range layers {
		router = layer.Value.Next(router)
	}
	router = http.NewCrossOriginProtection().Handler(router)

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
		slog.Info("Listening", "UDS", config.UDS)
		return http.Serve(ln, router)
	}
	slog.Info("Listening", "TCP", config.Address)
	slog.Info("Listening", "http", "http://"+config.Address)

	return http.ListenAndServe(config.Address, router)
}
