package p_announcements

import (
	"log"
	"time"

	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/p_filesystem"
	"github.com/lariv-in/lago/p_users"
	"gorm.io/gorm"
)

// Announcement represents an academic announcement.
//
// Source parity notes (Django):
// - release_at defaults to now
// - expiry_at is nullable
// - created_by is nullable and set to NULL on user deletion (best-effort parity)
type Announcement struct {
	gorm.Model

	Title       string `gorm:"notnull"`
	Description string `gorm:"type:text;notnull"`
	URL         string

	CreatedByID *uint
	CreatedBy   *p_users.User `gorm:"constraint:OnDelete:SET NULL;foreignKey:CreatedByID;references:ID"`

	ReleaseAt time.Time
	ExpiryAt  *time.Time

	Assets []p_filesystem.VNode `gorm:"many2many:announcement_assets;"`
}

func (a *Announcement) BeforeSave(tx *gorm.DB) error {
	// Django default for release_at.
	if a.ReleaseAt.IsZero() {
		a.ReleaseAt = time.Now()
	}
	return nil
}

func init() {
	lago.OnDBInit(func(d *gorm.DB) *gorm.DB {
		if err := d.AutoMigrate(&Announcement{}); err != nil {
			log.Panicf("failed to migrate Announcement model: %v", err)
		}
		return d
	})

	lago.RegistryAdmin.Register("p_announcements", lago.AdminPanel[Announcement]{
		SearchField: "Title",
		ListFields: []string{
			"Title",
			"Description",
			"URL",
			"ReleaseAt",
			"ExpiryAt",
			"CreatedBy.Name",
			"UpdatedAt",
		},
		Preload: []string{
			"CreatedBy",
		},
	})
}
