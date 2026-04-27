package p_announcements

import (
	"time"

	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_semesters"
	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/registry"
	"gorm.io/gorm"
)

// Announcement mirrors announcements.Announcement (Description = Django `description`).
type Announcement struct {
	gorm.Model

	Title       string
	Description string `gorm:"type:text"`
	IsUniversal bool   `gorm:"not null;default:false"`
	ReleaseAt   time.Time
	ExpiryAt    *time.Time
	Priority    string // 1 / 2 / 3
	SemesterID  *uint  `gorm:"index"`
	Semester    *p_semesters.Semester `gorm:"foreignKey:SemesterID"`
	CreatedByID *uint  `gorm:"index"`
	CreatedBy   *p_users.User          `gorm:"foreignKey:CreatedByID"`
	SignedByID  *uint  `gorm:"index"`
	SignedBy    *p_users.User          `gorm:"foreignKey:SignedByID"`
}

// AnnouncementPriorityChoices match Django-style numeric priority labels.
var AnnouncementPriorityChoices = []registry.Pair[string, string]{
	{Key: "1", Value: "1 — Low"},
	{Key: "2", Value: "2 — Normal"},
	{Key: "3", Value: "3 — High"},
}

func init() {
	lago.OnDBInit("p_announcements.models", func(d *gorm.DB) *gorm.DB {
		lago.RegisterModel[Announcement](d)
		return d
	})
	lago.RegistryAdmin.Register("p_announcements", lago.AdminPanel[Announcement]{
		SearchField: "Title",
		ListFields:  []string{"Title", "ReleaseAt", "ExpiryAt", "Priority", "SemesterID"},
	})
}
