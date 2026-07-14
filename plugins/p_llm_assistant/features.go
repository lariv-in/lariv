package p_llm_assistant

import (
	"github.com/lariv-in/lago"
	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/registry"
	"github.com/lariv-in/lago/views"
)

var (
	pluginPageEntries       []registry.Pair[string, components.PageInterface]
	pluginPagePatches       []registry.Pair[string, func(components.PageInterface) components.PageInterface]
	pluginViewEntries       []registry.Pair[string, *views.View]
	pluginViewPatches       []registry.Pair[string, func(*views.View) *views.View]
	pluginRouteEntries      []registry.Pair[string, lago.Route]
	pluginRoutePatches      []registry.Pair[string, func(lago.Route) lago.Route]
	pluginConfigEntries     []registry.Pair[string, lago.Config]
	pluginDBInitHookEntries []registry.Pair[string, lago.DBInitHook]
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

func registerPluginRoute(key string, value lago.Route) {
	pluginRouteEntries = append(pluginRouteEntries, registry.Pair[string, lago.Route]{Key: key, Value: value})
}

func patchPluginRoute(key string, patch func(lago.Route) lago.Route) {
	pluginRoutePatches = append(pluginRoutePatches, registry.Pair[string, func(lago.Route) lago.Route]{Key: key, Value: patch})
}

func registerPluginConfig(key string, value lago.Config) {
	pluginConfigEntries = append(pluginConfigEntries, registry.Pair[string, lago.Config]{Key: key, Value: value})
}

func registerPluginDBInitHook(key string, value lago.DBInitHook) {
	pluginDBInitHookEntries = append(pluginDBInitHookEntries, registry.Pair[string, lago.DBInitHook]{Key: key, Value: value})
}

func pluginPages() lago.PluginFeatures[components.PageInterface] {
	return lago.PluginFeatures[components.PageInterface]{Entries: pluginPageEntries, Patches: pluginPagePatches}
}

func pluginViews() lago.PluginFeatures[*views.View] {
	return lago.PluginFeatures[*views.View]{Entries: pluginViewEntries, Patches: pluginViewPatches}
}

func pluginRoutes() lago.PluginFeatures[lago.Route] {
	return lago.PluginFeatures[lago.Route]{Entries: pluginRouteEntries, Patches: pluginRoutePatches}
}

func pluginConfigs() lago.PluginFeatures[lago.Config] {
	return lago.PluginFeatures[lago.Config]{Entries: pluginConfigEntries}
}

func pluginDBInitHooks() lago.PluginFeatures[lago.DBInitHook] {
	return lago.PluginFeatures[lago.DBInitHook]{Entries: pluginDBInitHookEntries}
}
