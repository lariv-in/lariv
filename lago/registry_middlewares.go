package lago

import (
	"github.com/lariv-in/lago/registry"
	"github.com/lariv-in/lago/views"
)

var RegistryMiddleware registry.Registry[views.Middleware] = registry.NewRegistry[views.Middleware]()
