package p_nirmancampus_courses

import (
	"github.com/lariv-in/lago/lago"
	"gorm.io/gorm"
)

type Course struct {
	gorm.Model

	Name        string
	IsActive    bool
	Level       string
	Code        string
	Subject     string
	Description string
}

func init() {
	lago.OnDBInit(func(d *gorm.DB) *gorm.DB {
		d.AutoMigrate(Course{})
		return d
	})
	lago.RegistryAdmin.Register("p_nirmancampus_courses", lago.AdminPanel[Course]{SearchField: "Name"})
}
