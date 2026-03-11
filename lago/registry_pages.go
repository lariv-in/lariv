package lago

import (
	"github.com/lariv-in/components"
	"github.com/lariv-in/registry"
)

var RegistryPage registry.Registry[components.PageInterface] = registry.NewRegistry[components.PageInterface]()
