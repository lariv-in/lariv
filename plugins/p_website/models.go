package p_website

import (
	"time"

	"github.com/lariv-in/lariv"
	"github.com/lariv-in/lariv/plugins/p_filesystem"
	"github.com/lariv-in/lariv/registry"
	"gorm.io/gorm"
)

// DBRoute represents a dynamic website route mapping a URL path to a page template in the filesystem.
type DBRoute struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	Path      string             `gorm:"uniqueIndex;notnull"`
	LTreePath string             `gorm:"column:ltree_path;type:ltree;->"`
	PageID    uint               `gorm:"notnull"`
	Page       p_filesystem.VNode   `gorm:"constraint:OnDelete:RESTRICT;foreignKey:PageID;references:ID"`
	References []p_filesystem.VNode `gorm:"many2many:p_website_route_references;"`
	IsActive   bool                 `gorm:"notnull;default:true"`
	// Theme is a GrapesJSThemes registry key (e.g. "p_website.default"); empty means no theme.
	Theme string `gorm:"type:varchar(128);not null;default:''"`
	// GrapesProject holds GrapesJS project JSON for re-editing; empty means never saved from the builder.
	GrapesProject string `gorm:"type:text"`
}

func pluginModels() lariv.PluginFeatures[any] {
	return lariv.PluginFeatures[any]{
		Entries: []registry.Pair[string, any]{
			{Key: "p_website.DBRoute", Value: DBRoute{}},
		},
	}
}

func pluginAdmin() lariv.PluginFeatures[lariv.AdminPanelInterface] {
	return lariv.PluginFeatures[lariv.AdminPanelInterface]{
		Entries: []registry.Pair[string, lariv.AdminPanelInterface]{
			{Key: "p_website", Value: lariv.AdminPanel[DBRoute]{
				SearchField: "Path",
				ListFields:  []string{"Path", "LTreePath", "PageID", "IsActive", "UpdatedAt"},
			}},
		},
	}
}
