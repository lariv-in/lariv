package lago

import (
	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/registry"
)

var RegistryPage *registry.Registry[components.PageInterface] = registry.NewRegistry[components.PageInterface]()
