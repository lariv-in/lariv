package p_livereloading

import (
	"io"

	"github.com/lariv-in/lago/lago"

	"golang.org/x/net/websocket"
)

func NilServer(ws *websocket.Conn) {
	io.Copy(io.Discard, ws)
}

func registerRoutes() {
	_ = lago.RegistryRoute.Register("livereloading.ws", lago.Route{
		Path:    "/_livereload",
		Handler: websocket.Handler(NilServer),
	})

}

func init() {
	registerRoutes()
}
