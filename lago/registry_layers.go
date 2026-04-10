package lago

import (
	"github.com/lariv-in/lago/registry"
	"github.com/lariv-in/lago/views"
)

var RegistryLayer *registry.Registry[views.GlobalLayer] = registry.NewRegistry[views.GlobalLayer]()
