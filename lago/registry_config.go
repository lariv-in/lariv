package lago

import (
	"github.com/lariv-in/lago/registry"
)

type Config interface {
	PostConfig()
}

var RegistryConfig *registry.Registry[Config] = registry.NewRegistry[Config]()
