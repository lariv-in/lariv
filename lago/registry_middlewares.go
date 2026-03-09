package lago

import (
	"net/http"
)

type Middleware = func (http.Handler) http.Handler

var RegistryMiddleware Registry[Middleware] = NewRegistry[Middleware]()
