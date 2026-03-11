package lago

import (
	"net/url"

	"github.com/lariv-in/registry"
)

type PluginType int

const (
	// For plugins that add new models and functionality, ideally independent of other plugins
	PluginTypeApp = iota
	// For plugins that add additional functionality to App
	PluginTypeAddon
	// For plugins that add a long running service
	PluginTypeService
)

type Plugin struct {
	Type        PluginType
	Icon        string
	Url         *url.URL
	VerboseName string
	Method      string
	OnClick     string
	Classes     string
	RenderKeys  []string
}

var RegistryPlugins registry.Registry[Plugin] = registry.NewRegistry[Plugin]()
