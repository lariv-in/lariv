package p_blog

import (
	"strings"

	"github.com/lariv-in/lariv"
	"github.com/lariv-in/lariv/getters"
	"github.com/lariv-in/lariv/plugins/p_users"
	"github.com/lariv-in/lariv/registry"
	"gorm.io/gorm"
)

// Blog represents a blog article in the system.
type Blog struct {
	gorm.Model

	Title       string       `gorm:"notnull"`
	Slug        string       `gorm:"type:varchar(255);uniqueIndex;notnull"`
	Description string       `gorm:"type:text"`
	CreatedByID uint         `gorm:"notnull"`
	CreatedBy   p_users.User `gorm:"constraint:OnDelete:CASCADE;foreignKey:CreatedByID"`
	Content     string       `gorm:"type:text"`
	Tags        []BlogTag    `gorm:"many2many:p_blog_tags;"`
}

// BeforeSave GORM hook to automatically generate a slug from Title if empty or when requested.
func (b *Blog) BeforeSave(tx *gorm.DB) error {
	if strings.TrimSpace(b.Slug) == "" {
		if strings.TrimSpace(b.Title) != "" {
			b.Slug = getters.TitleToFormSlug(b.Title)
		} else {
			b.Slug = "blog"
		}
	} else {
		b.Slug = getters.TitleToFormSlug(b.Slug)
	}
	return nil
}

// BlogTag represents a hierarchical blog tag using ltree.
type BlogTag struct {
	gorm.Model

	Name  string `gorm:"type:ltree;notnull"`
	Blogs []Blog `gorm:"many2many:p_blog_tags;"`
}

func pluginModels() lariv.PluginFeatures[any] {
	return lariv.PluginFeatures[any]{
		Entries: []registry.Pair[string, any]{
			{Key: "p_blog.Blog", Value: Blog{}},
			{Key: "p_blog.BlogTag", Value: BlogTag{}},
		},
	}
}

func init() {
	lariv.RegistryAdmin.Register("p_blog", lariv.AdminPanel[Blog]{
		SearchField: "Title",
		ListFields:  []string{"Title", "Slug", "CreatedByID", "UpdatedAt"},
		Preload:     []string{"CreatedBy", "Tags"},
	})
	lariv.RegistryAdmin.Register("p_blog_tag", lariv.AdminPanel[BlogTag]{
		SearchField: "Name",
		ListFields:  []string{"Name", "UpdatedAt"},
	})
}
