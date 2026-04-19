package p_seer_websites

import (
	"github.com/lariv-in/lago/lago"
	"gorm.io/gorm"
)

const WebsitesTable = "seer_websites"

// Website stores a scraped page as markdown plus the canonical URL.
type Website struct {
	gorm.Model

	URL      lago.PageURL `gorm:"column:url;type:text;not null;default:''"`
	Markdown string  `gorm:"type:text;not null;default:''"`
}

func (Website) TableName() string {
	return WebsitesTable
}

func init() {
	lago.OnDBInit("p_seer_websites.models", func(db *gorm.DB) *gorm.DB {
		lago.RegisterModel[Website](db)
		return db
	})
}
