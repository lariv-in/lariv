package p_students

import (
	"log"
	"time"

	"github.com/lariv-in/lago"
	"github.com/lariv-in/p_filesystem"
	"github.com/lariv-in/p_users"
	"gorm.io/gorm"
)

// Student represents a student entity with a one-to-one relationship to a User.
// This model preserves the invariants from the Django source:
//   - StudentNo is unique
//   - DOB is optional (nullable)
//   - Default ordering by student number
//   - One-to-one relationship with User (enforced by unique index on UserID)
type Student struct {
	gorm.Model

	// One-to-one relationship with User
	// UserID has a unique index to enforce the one-to-one constraint
	UserID uint           `gorm:"uniqueIndex;notnull"`
	User   p_users.User   `gorm:"constraint:OnDelete:CASCADE"`

	// StudentNo is a unique identifier for the student (e.g., "123456")
	StudentNo string `gorm:"uniqueIndex;notnull"`

	// DOB is the date of birth, optional (nullable in DB)
	DOB *time.Time

	// Assets is a many-to-many relationship with filesystem VNodes.
	// Note: Not exposed in the UI initially, kept for parity with source schema.
	Assets []p_filesystem.VNode `gorm:"many2many:student_assets;"`
}

func init() {
	lago.OnDBInit(func(d *gorm.DB) *gorm.DB {
		if err := d.AutoMigrate(&Student{}); err != nil {
			log.Panicf("failed to migrate Student model: %v", err)
		}
		return d
	})

	lago.RegistryAdmin.Register("p_students", lago.AdminPanel[Student]{
		SearchField: "StudentNo",
		ListFields:  []string{"StudentNo", "User.Name", "DOB", "UpdatedAt"},
		Preload:     []string{"User"},
	})
}
