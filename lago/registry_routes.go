package lago

import (
	"context"
	"io"
	"net/http"

	"github.com/lariv-in/getters"
	"github.com/lariv-in/registry"
	"golang.org/x/net/websocket"
)

var RegistryRoute registry.Registry[Route] = registry.NewRegistry[Route]()

type Route struct {
	Path    string
	Handler http.Handler
}

// Echo the data received on the WebSocket.
func EchoServer(ws *websocket.Conn) {
	io.Copy(ws, ws)
}

func GetRouter() *http.ServeMux {
	baseRouter := http.NewServeMux()
	routes := RegistryRoute.All()
	for _, route := range *routes {
		baseRouter.Handle(route.Path+"{$}", route.Handler)
	}
	return baseRouter
}

// RoutePathGetter returns a Getter that resolves to the route's Path string.
func RoutePathGetter(name string) getters.Getter {
	return func(ctx context.Context) any {
		if route, ok := RegistryRoute.Get(name); ok {
			return route.Path
		}
		return nil
	}
}
