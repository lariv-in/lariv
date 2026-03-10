package lago

import (
	"net/http"
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
		baseRouter.Handle(route.Path + "{$}", route.Handler)
	}
	return baseRouter
}
