package p_livereloading

import (
	"github.com/lariv-in/lariv/components"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func liveReloadScript() string {
	return "(function() {" +
		"const host = window.location.hostname;" +
		"const allowedHosts = " + allowedHostsJS() + ";" +
		"if (!allowedHosts.includes(host)) {" +
		"return;" +
		"}" +
		"let isReconnecting = false;" +
		"const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';" +
		"" +
		"function connectLiveReload() {" +
		"try {" +
		"const ws = new WebSocket(`${protocol}//${window.location.host}/_livereload`);" +
		"" +
		"ws.onclose = function() {" +
		"if (!isReconnecting) {" +
		"console.warn('Live Reload: Server disconnected. Polling for restart...');" +
		"isReconnecting = true;" +
		"}" +
		"setTimeout(connectLiveReload, 100); " +
		"};" +
		"" +
		"ws.onopen = function() {" +
		"if (isReconnecting) {" +
		"console.log('Live Reload: Server is back. Reloading page...');" +
		"window.location.reload();" +
		"}" +
		"};" +
		"ws.onerror = function() {" +
		"ws.close(); " +
		"};" +
		"} catch (err) {" +
		"console.warn('Live Reload: unavailable', err);" +
		"}" +
		"}" +
		"" +
		"connectLiveReload();" +
		"})();"
}

func registerHeadNodes() {
	components.RegistryShellHeadNodes.Register(
		"liverealoading.js",
		Script(Raw(liveReloadScript())),
	)
}

func init() {
	registerHeadNodes()
}
