package p_dashboard

import "github.com/lariv-in/lago"

func pluginStages[T any](stage func() lago.PluginFeatures[T]) []func() lago.PluginFeatures[T] {
	return []func() lago.PluginFeatures[T]{stage}
}
