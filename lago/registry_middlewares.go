package lago

import (
	"github.com/lariv-in/lago/registry"
	"github.com/lariv-in/lago/views"
)

var RegistryMiddleware registry.Registry[views.GlobalMiddleware] = registry.NewRegistry[views.GlobalMiddleware]()
