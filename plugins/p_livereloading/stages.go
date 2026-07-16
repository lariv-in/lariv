package p_livereloading

import "github.com/lariv-in/lariv"

func pluginStages[T any](stage func() lariv.PluginFeatures[T]) []func() lariv.PluginFeatures[T] {
	return []func() lariv.PluginFeatures[T]{stage}
}
