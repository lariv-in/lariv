package p_livereloading

import (
	"context"
	"io"

	"github.com/lariv-in/lariv"
	"github.com/lariv-in/lariv/components"
	"github.com/lariv-in/lariv/registry"
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

type liveReloadHead struct {
	components.Page
}

func (liveReloadHead) Build(cat components.Catalog, ctx context.Context, w io.Writer) error {
	_, err := io.WriteString(w, "<script>"+liveReloadScript()+"</script>")
	return err
}

func pluginHeadNodes() lariv.PluginFeatures[components.PageInterface] {
	return lariv.PluginFeatures[components.PageInterface]{
		Entries: []registry.Pair[string, components.PageInterface]{
			{Key: "liverealoading.js", Value: liveReloadHead{}},
		},
	}
}
