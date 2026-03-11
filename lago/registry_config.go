package lago

import (
	"github.com/lariv-in/registry"
)

type Config interface {
	PostConfig()
}

var RegistryConfig registry.Registry[Config] = registry.NewRegistry[Config]()
