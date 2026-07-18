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
	Page      p_filesystem.VNode `gorm:"constraint:OnDelete:RESTRICT;foreignKey:PageID;references:ID"`
	IsActive  bool               `gorm:"notnull;default:true"`
}

func pluginModels() lariv.PluginFeatures[any] {
	return lariv.PluginFeatures[any]{
		Entries: []registry.Pair[string, any]{
			{Key: "p_website.DBRoute", Value: DBRoute{}},
		},
	}
}

func init() {
	lariv.RegistryAdmin.Register("p_website", lariv.AdminPanel[DBRoute]{
		SearchField: "Path",
		ListFields:  []string{"Path", "LTreePath", "PageID", "IsActive", "UpdatedAt"},
	})
}
