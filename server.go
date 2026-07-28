package lariv

import "fmt"

// StartServer is deprecated. Use (*App).StartServer after [AppBuilder.LoadConfigFromFile].
func StartServer(config LarivConfig) error {
	_ = config
	return fmt.Errorf("lariv.StartServer is removed; use (*App).StartServer")
}
