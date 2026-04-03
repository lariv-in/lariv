package p_nirmancampus_website

import (

	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_filesystem"
	"gorm.io/gorm"
)

type StudentZoneSection struct {
	gorm.Model

	Title string `gorm:"notnull"`
	Order int    `gorm:"notnull;default:0"`
}

type StudentZoneItem struct {
	gorm.Model

	Title string `gorm:"notnull"`

	IsLink bool
	Link   string

	FileID *uint
	File   *p_filesystem.VNode `gorm:"constraint:OnDelete:SET NULL;foreignKey:FileID;references:ID"`

	StudentZoneSectionID uint
	StudentZoneSection   StudentZoneSection `gorm:"constraint:OnDelete:CASCADE;foreignKey:StudentZoneSectionID;references:ID"`
}

func init() {
	lago.OnDBInit(func(d *gorm.DB) *gorm.DB {
		lago.RegisterModel[StudentZoneSection](d)
		lago.RegisterModel[StudentZoneItem](d)
		return d
	})

	lago.RegistryAdmin.Register("p_nirmancampus_website.student_zone_sections", lago.AdminPanel[StudentZoneSection]{
		SearchField: "Title",
		ListFields:  []string{"Title", "Order", "UpdatedAt"},
	})

	lago.RegistryAdmin.Register("p_nirmancampus_website.student_zone_items", lago.AdminPanel[StudentZoneItem]{
		SearchField: "Title",
		ListFields:  []string{"Title", "IsLink", "Link", "StudentZoneSection.Title", "UpdatedAt"},
		Preload:     []string{"StudentZoneSection", "File"},
	})
}
