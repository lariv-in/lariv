package p_teachers

import (
	"log"

	"github.com/lariv-in/lago"
	"github.com/lariv-in/p_filesystem"
	"github.com/lariv-in/p_users"
	"gorm.io/gorm"
)

// Teacher represents a teacher entity with a one-to-one relationship to a User.
// This model preserves the invariants from the Django source:
//   - Code is unique
//   - One-to-one relationship with User (enforced by unique index on UserID)
//   - ProfilePhoto is nullable FK to VNode
//   - Assets is many-to-many relationship with VNode (not exposed in UI initially)
//   - Default ordering by Code
type Teacher struct {
	gorm.Model

	// One-to-one relationship with User
	// UserID has a unique index to enforce the one-to-one constraint
	UserID uint         `gorm:"uniqueIndex;notnull"`
	User   p_users.User `gorm:"constraint:OnDelete:CASCADE"`

	// ProfilePhoto is a nullable FK to a filesystem VNode (optional profile photo)
	ProfilePhotoID *uint               `gorm:"index"`
	ProfilePhoto   *p_filesystem.VNode `gorm:"constraint:OnDelete:SET NULL"`

	// Code is a unique identifier for the teacher (e.g., "TCH001")
	Code string `gorm:"uniqueIndex;notnull"`

	// Qualifications is a text field describing teacher qualifications (nullable)
	Qualifications *string

	// Assets is a many-to-many relationship with filesystem VNodes.
	// Note: Not exposed in the UI initially, kept for parity with source schema.
	Assets []p_filesystem.VNode `gorm:"many2many:teacher_assets;"`
}

func init() {
	lago.OnDBInit(func(d *gorm.DB) *gorm.DB {
		if err := d.AutoMigrate(&Teacher{}); err != nil {
			log.Panicf("failed to migrate Teacher model: %v", err)
		}
		return d
	})

	lago.RegistryAdmin.Register("p_teachers", lago.AdminPanel[Teacher]{
		SearchField: "Code",
		ListFields:  []string{"Code", "User.Name", "Qualifications", "UpdatedAt"},
		Preload:     []string{"User"},
	})
}
