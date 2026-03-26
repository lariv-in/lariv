package p_nirmancampus_courses

import (
	"log"

	"github.com/lariv-in/lago/lago"
	"gorm.io/gorm"
)

type Course struct {
	gorm.Model

	Name        string
	IsActive    bool
	Code        string
	Subject     string
	Description string
}

func init() {
	lago.OnDBInit(func(d *gorm.DB) *gorm.DB {
		if err := d.AutoMigrate(&Course{}); err != nil {
			log.Panicf("failed to migrate Course model: %v", err)
		}
		return d
	})
	lago.RegistryAdmin.Register("p_nirmancampus_courses", lago.AdminPanel[Course]{SearchField: "Name"})
}
