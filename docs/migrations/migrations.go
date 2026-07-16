// Package migrations contains explanations and code examples for database migrations in Lariv.
//
// # Database Migrations (migrations.go)
//
// Database schema migrations in Lariv are managed sequentially using SQL scripts compatible with the goose engine.
// Migrations are isolated per plugin, embedded using Go's go:embed directive, and registered within the Plugin struct.
//
// # Generating SQL Migration Files
//
// To create a new goose SQL migration file in the plugin's migrations folder, execute the goose CLI command from the repository root:
//
//	goose -dir plugins/<plugin_name>/migrations create <migration_name> sql
//
// Example:
//
//	goose -dir plugins/blog/migrations create create_posts_table sql
//
// This produces a SQL script (e.g. plugins/blog/migrations/20260713161200_create_posts_table.sql) with:
//
//	-- +goose Up
//	CREATE TABLE posts ( ... );
//
//	-- +goose Down
//	DROP TABLE posts;
//
// # Example Migrations File
//
//	package myplugin
//
//	import (
//		"embed"
//		"github.com/lariv-in/lariv"
//		"github.com/lariv-in/lariv/registry"
//	)
//
//	//go:embed migrations
//	var migrationsFS embed.FS
//
//	func pluginMigrations() lariv.PluginFeatures[lariv.UsefulFilesystem] {
//		return lariv.PluginFeatures[lariv.UsefulFilesystem]{
//			Entries: []registry.Pair[string, lariv.UsefulFilesystem]{
//				{Key: "myplugin.migrations", Value: migrationsFS},
//			},
//		}
//	}
package migrations
