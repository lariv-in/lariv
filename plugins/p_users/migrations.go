package p_users

import (
	"embed"

	"github.com/lariv-in/lago"
	"github.com/lariv-in/lago/registry"
)

//go:embed migrations
var migrationsFS embed.FS

func pluginMigrations() lago.PluginFeatures[lago.UsefulFilesystem] {
	return lago.PluginFeatures[lago.UsefulFilesystem]{
		Entries: []registry.Pair[string, lago.UsefulFilesystem]{
			{Key: "p_users.migrations", Value: migrationsFS},
		},
	}
}
