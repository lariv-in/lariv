package p_livereloading

import (
	"io"

	"github.com/lariv-in/lago"
	"github.com/lariv-in/lago/registry"

	"golang.org/x/net/websocket"
)

func NilServer(ws *websocket.Conn) {
	io.Copy(io.Discard, ws)
}

func pluginRoutes() lago.PluginFeatures[lago.Route] {
	return lago.PluginFeatures[lago.Route]{
		Entries: []registry.Pair[string, lago.Route]{
			{
				Key: "livereloading.ws",
				Value: lago.Route{
					Path:    "/_livereload",
					Handler: websocket.Handler(NilServer),
				},
			},
		},
	}
}
