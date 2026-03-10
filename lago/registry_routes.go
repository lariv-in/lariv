package lago

import (
	"context"
	"net/http"

	"github.com/lariv-in/components"
)

var RegistryRoute Registry[Route] = NewRegistry[Route]()

type Route struct {
	Path    string
	Handler http.Handler
}

func GetRouter() *http.ServeMux {
	baseRouter := http.NewServeMux()
	routes := RegistryRoute.All()
	for _, route := range *routes {
		baseRouter.Handle(route.Path+"{$}", route.Handler)
	}
	return baseRouter
}

// PathGetter returns a Getter that resolves to the route's Path string.
func RoutePathGetter(name string) components.Getter {
	return func(ctx context.Context) any {
		if route, ok := RegistryRoute.Get(name); ok {
			return route.Path
		}
		return nil
	}
}
