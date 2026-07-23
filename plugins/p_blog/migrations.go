package p_blog

import (
	"embed"

	"github.com/lariv-in/lariv"
	"github.com/lariv-in/lariv/registry"
)

//go:embed migrations
var migrationsFS embed.FS

func pluginMigrations() lariv.PluginFeatures[lariv.UsefulFilesystem] {
	return lariv.PluginFeatures[lariv.UsefulFilesystem]{
		Entries: []registry.Pair[string, lariv.UsefulFilesystem]{
			{Key: "p_blog.migrations", Value: migrationsFS},
		},
	}
}
