// Package models contains explanations and code examples for database models in Lariv.
//
// # Database Models (models.go)
//
// Database schemas in Lariv are defined as GORM models and auto-migrated on startup.
// Define your database entities as structs and register them inside the plugin.
//
// # Model Definition and Registration Example
//
//	package myplugin
//
//	import (
//		"time"
//		"github.com/lariv-in/lariv"
//		"github.com/lariv-in/lariv/registry"
//	)
//
//	type BlogPost struct {
//		ID        uint      `gorm:"primaryKey"`
//		Title     string    `gorm:"size:255;not null"`
//		Body      string    `gorm:"type:text"`
//		CreatedAt time.Time
//		UpdatedAt time.Time
//	}
//
//	func pluginModels() lariv.PluginFeatures[any] {
//		return lariv.PluginFeatures[any]{
//			Entries: []registry.Pair[string, any]{
//				{Key: "blog.post", Value: &BlogPost{}},
//			},
//		}
//	}
//
//	// Registering inside lariv.Plugin
//	lariv.Plugin{
//		Models: lariv.PluginStages(pluginModels),
//	}
//
// # GORM Reference
//
// For details on database tag properties, associations, and queries, refer to the [gorm.io/gorm] package documentation.
package models
