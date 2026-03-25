package p_assignments

import (
	"log"
	"time"

	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_filesystem"
	"gorm.io/gorm"
)

// Assignment is a coursework item with a due datetime, optional attachments, and a maximum score.
type Assignment struct {
	gorm.Model

	Name        string    `gorm:"notnull"`
	Due         time.Time `gorm:"notnull"`
	Description string    `gorm:"type:text"`
	MaxMarks    int       `gorm:"notnull"`

	Assets []p_filesystem.VNode `gorm:"many2many:assignment_assets;"`
}

func init() {
	lago.OnDBInit(func(d *gorm.DB) *gorm.DB {
		if err := d.AutoMigrate(&Assignment{}); err != nil {
			log.Panicf("failed to migrate Assignment model: %v", err)
		}
		return d
	})

	lago.RegistryAdmin.Register("p_assignments", lago.AdminPanel[Assignment]{
		SearchField: "Name",
		ListFields: []string{
			"Name",
			"Due",
			"MaxMarks",
			"Description",
			"UpdatedAt",
		},
		Preload: []string{"Assets"},
	})
}
