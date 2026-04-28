package p_events

import (
	"time"

	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_semesters"
	"gorm.io/gorm"
)

// SchoolEvent mirrors events.Event.
type SchoolEvent struct {
	gorm.Model

	Title       string
	Description string `gorm:"type:text"`
	StartsAt    time.Time
	EndsAt      *time.Time
	IsUniversal bool  `gorm:"not null;default:true"`
	IsActive    bool  `gorm:"not null;default:true"`
	SemesterID  *uint `gorm:"index"`
	Semester    *p_semesters.Semester `gorm:"foreignKey:SemesterID"`
}

func init() {
	lago.OnDBInit("p_events.models", func(d *gorm.DB) *gorm.DB {
		lago.RegisterModel[SchoolEvent](d)
		return d
	})
	lago.RegistryAdmin.Register("p_events", lago.AdminPanel[SchoolEvent]{
		SearchField: "Title",
		ListFields:  []string{"Title", "StartsAt", "EndsAt", "IsActive", "SemesterID"},
	})
}
