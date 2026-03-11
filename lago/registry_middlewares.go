package lago

import (
	"net/http"

	"github.com/lariv-in/registry"
)

type Middleware = func (http.Handler) http.Handler

var RegistryMiddleware registry.Registry[Middleware] = registry.NewRegistry[Middleware]()
