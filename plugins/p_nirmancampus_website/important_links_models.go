package p_nirmancampus_website

import (
	"log"

	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/p_filesystem"
	"gorm.io/gorm"
)

type ImportantLink struct {
	gorm.Model

	Title string `gorm:"notnull"`
	Order int    `gorm:"notnull;default:0"`

	// If IsLink is true, open Link as a normal URL.
	// If IsLink is false, download the attached File.
	IsLink bool
	Link   string

	FileID *uint
	File   *p_filesystem.VNode `gorm:"constraint:OnDelete:SET NULL;foreignKey:FileID;references:ID"`
}

func init() {
	lago.OnDBInit(func(d *gorm.DB) *gorm.DB {
		if err := d.AutoMigrate(&ImportantLink{}); err != nil {
			log.Panicf("failed to migrate ImportantLink model: %v", err)
		}
		return d
	})

	lago.RegistryAdmin.Register("p_nirmancampus_website.important_links", lago.AdminPanel[ImportantLink]{
		SearchField: "Title",
		ListFields:  []string{"Title", "Order", "IsLink", "Link", "UpdatedAt"},
		Preload:     []string{"File"},
	})
}

