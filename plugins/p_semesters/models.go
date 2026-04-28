package p_semesters

import (
	"time"

	"github.com/lariv-in/lago/lago"
	"gorm.io/gorm"
)

// Semester mirrors semesters.Semester (Django `start` / `end` → Start / End).
type Semester struct {
	gorm.Model

	Code     string `gorm:"uniqueIndex"`
	Name     string `gorm:"not null"`
	Start    time.Time
	End      time.Time
	IsActive bool `gorm:"not null;default:true"`
}

func init() {
	lago.OnDBInit("p_semesters.models", func(d *gorm.DB) *gorm.DB {
		lago.RegisterModel[Semester](d)
		return d
	})
	lago.RegistryAdmin.Register("p_semesters", lago.AdminPanel[Semester]{
		SearchField: "Name",
		ListFields:  []string{"Code", "Name", "Start", "End", "IsActive"},
	})
}
