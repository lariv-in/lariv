package lago

import (
	"context"
	"net/http"

	"github.com/lariv-in/components"
)

func GetterRoute(routeKey string) components.Getter {
	return func(ctx context.Context) any {
		route, ok := RegistryRoute.Get(routeKey)
		if !ok {
			return ""
		}
		return route.Path
	}
}

var RegistryRoute Registry[Route] = NewRegistry[Route]()

type Route struct {
	Path    string
	Handler http.Handler
}

func GetRouter() *http.ServeMux {
	baseRouter := http.NewServeMux()
	routes := RegistryRoute.All()
	for _, route := range *routes {
		baseRouter.Handle(route.Path, route.Handler)
	}
	return baseRouter
}
