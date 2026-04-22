package p_livereloading

import (
	"github.com/lariv-in/lago/components"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func registerHeadNodes() {
	components.RegistryShellHeadNodes.Register("liverealoading.js",
		Script(Raw("(function() {"+
			"let isReconnecting = false;"+
			""+
			"function connectLiveReload() {"+
			"const ws = new WebSocket(`ws://${window.location.host}/_livereload`);"+
			""+
			"ws.onclose = function() {"+
			"if (!isReconnecting) {"+
			"console.warn('Live Reload: Server disconnected. Polling for restart...');"+
			"isReconnecting = true;"+
			"}"+
			"setTimeout(connectLiveReload, 100); "+
			"};"+
			""+
			"ws.onopen = function() {"+
			"if (isReconnecting) {"+
			"console.log('Live Reload: Server is back. Reloading page...');"+
			"window.location.reload();"+
			"}"+
			"};"+
			"ws.onerror = function(err) {"+
			"ws.close(); "+
			"};"+
			"}"+
			""+
			"connectLiveReload();"+
			"})();")),
	)
}

func init() {
	registerHeadNodes()
}
