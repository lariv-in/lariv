package lago

import (
	"net/url"

	"github.com/lariv-in/lago/registry"
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
	URL         *url.URL
	VerboseName string
	Method      string
	OnClick     string
	Classes     string
	Roles       []string
}

var RegistryPlugin *registry.Registry[Plugin] = registry.NewRegistry[Plugin]()
