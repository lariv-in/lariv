package p_llm_assistant

import (
	"github.com/lariv-in/lariv"
	"github.com/lariv-in/lariv/components"
	"github.com/lariv-in/lariv/registry"
	"github.com/lariv-in/lariv/views"
)

var (
	pluginPageEntries       []registry.Pair[string, components.PageInterface]
	pluginPagePatches       []registry.Pair[string, func(components.PageInterface) components.PageInterface]
	pluginViewEntries       []registry.Pair[string, *views.View]
	pluginViewPatches       []registry.Pair[string, func(*views.View) *views.View]
	pluginRouteEntries      []registry.Pair[string, lariv.Route]
	pluginRoutePatches      []registry.Pair[string, func(lariv.Route) lariv.Route]
	pluginConfigEntries     []registry.Pair[string, lariv.Config]
	pluginDBInitHookEntries []registry.Pair[string, lariv.DBInitHook]
)

func registerPluginPage(key string, value components.PageInterface) {
	pluginPageEntries = append(pluginPageEntries, registry.Pair[string, components.PageInterface]{Key: key, Value: value})
}

func patchPluginPage(key string, patch func(components.PageInterface) components.PageInterface) {
	pluginPagePatches = append(pluginPagePatches, registry.Pair[string, func(components.PageInterface) components.PageInterface]{Key: key, Value: patch})
}

func registerPluginView(key string, value *views.View) {
	pluginViewEntries = append(pluginViewEntries, registry.Pair[string, *views.View]{Key: key, Value: value})
}

func patchPluginView(key string, patch func(*views.View) *views.View) {
	pluginViewPatches = append(pluginViewPatches, registry.Pair[string, func(*views.View) *views.View]{Key: key, Value: patch})
}

func registerPluginRoute(key string, value lariv.Route) {
	pluginRouteEntries = append(pluginRouteEntries, registry.Pair[string, lariv.Route]{Key: key, Value: value})
}

func patchPluginRoute(key string, patch func(lariv.Route) lariv.Route) {
	pluginRoutePatches = append(pluginRoutePatches, registry.Pair[string, func(lariv.Route) lariv.Route]{Key: key, Value: patch})
}

func registerPluginConfig(key string, value lariv.Config) {
	pluginConfigEntries = append(pluginConfigEntries, registry.Pair[string, lariv.Config]{Key: key, Value: value})
}

func registerPluginDBInitHook(key string, value lariv.DBInitHook) {
	pluginDBInitHookEntries = append(pluginDBInitHookEntries, registry.Pair[string, lariv.DBInitHook]{Key: key, Value: value})
}

func pluginPages() lariv.PluginFeatures[components.PageInterface] {
	return lariv.PluginFeatures[components.PageInterface]{Entries: pluginPageEntries, Patches: pluginPagePatches}
}

func pluginViews() lariv.PluginFeatures[*views.View] {
	return lariv.PluginFeatures[*views.View]{Entries: pluginViewEntries, Patches: pluginViewPatches}
}

func pluginRoutes() lariv.PluginFeatures[lariv.Route] {
	return lariv.PluginFeatures[lariv.Route]{Entries: pluginRouteEntries, Patches: pluginRoutePatches}
}

func pluginConfigs() lariv.PluginFeatures[lariv.Config] {
	return lariv.PluginFeatures[lariv.Config]{Entries: pluginConfigEntries}
}

func pluginDBInitHooks() lariv.PluginFeatures[lariv.DBInitHook] {
	return lariv.PluginFeatures[lariv.DBInitHook]{Entries: pluginDBInitHookEntries}
}
