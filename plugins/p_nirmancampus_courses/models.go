package p_nirmancampus_courses

import (

	"github.com/lariv-in/lago/lago"
	"gorm.io/gorm"
)

type Course struct {
	gorm.Model

	Name        string
	IsActive    bool
	Code        string
	CourseType  string `gorm:"type:varchar(64);not null;default:''"`
	Description string
}

func init() {
	lago.OnDBInit(func(d *gorm.DB) *gorm.DB {
		lago.RegisterModel[Course](d)
		return d
	})
	lago.RegistryAdmin.Register("p_nirmancampus_courses", lago.AdminPanel[Course]{SearchField: "Name"})
}
