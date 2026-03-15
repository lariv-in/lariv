package lago

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/lariv-in/getters"
	"github.com/lariv-in/registry"
)

var RegistryRoute registry.Registry[Route] = registry.NewRegistry[Route]()

type Route struct {
	Path    string
	Handler http.Handler
}

func GetRouter() *http.ServeMux {
	baseRouter := http.NewServeMux()
	routes := RegistryRoute.All()
	for _, route := range routes {
		baseRouter.Handle(route.Path+"{$}", route.Handler)
	}
	return baseRouter
}

// GetterRoutePath returns a Getter that resolves to the route's Path string.
func GetterRoutePath(name string, args map[string]getters.Getter) getters.Getter {
	return func(ctx context.Context) any {
		if route, ok := RegistryRoute.Get(name); ok {
			r := route.Path
			for k, g := range args {
				r = strings.ReplaceAll(r, fmt.Sprintf("{%s}", k), fmt.Sprintf("%v", g(ctx)))
			}
			return r
		}
		return nil
	}
}
