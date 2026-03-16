package lago

import (
	"github.com/lariv-in/registry"
	"github.com/lariv-in/views"
)

var RegistryMiddleware registry.Registry[views.Middleware] = registry.NewRegistry[views.Middleware]()
