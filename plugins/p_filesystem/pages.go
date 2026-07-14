package p_filesystem

import (
	"github.com/lariv-in/lago"
	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/registry"
)

func pluginPages() lago.PluginFeatures[components.PageInterface] {
	var entries []registry.Pair[string, components.PageInterface]
	entries = append(entries, pageEntriesMenus()...)
	entries = append(entries, pageEntriesFilters()...)
	entries = append(entries, pageEntriesTables()...)
	entries = append(entries, pageEntriesDetail()...)
	entries = append(entries, pageEntriesForms()...)
	entries = append(entries, pageEntriesSelection()...)
	return lago.PluginFeatures[components.PageInterface]{Entries: entries}
}
