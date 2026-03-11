package p_courses

import (
	"github.com/lariv-in/lago"
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
}

