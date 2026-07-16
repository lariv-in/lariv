package p_livereloading

import (
	"io"

	"github.com/lariv-in/lariv"
	"github.com/lariv-in/lariv/registry"

	"golang.org/x/net/websocket"
)

func NilServer(ws *websocket.Conn) {
	io.Copy(io.Discard, ws)
}

func pluginRoutes() lariv.PluginFeatures[lariv.Route] {
	return lariv.PluginFeatures[lariv.Route]{
		Entries: []registry.Pair[string, lariv.Route]{
			{
				Key: "livereloading.ws",
				Value: lariv.Route{
					Path:    "/_livereload",
					Handler: websocket.Handler(NilServer),
				},
			},
		},
	}
}
