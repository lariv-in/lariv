package p_students

import (
	"log"
	"time"

	"github.com/lariv-in/lago"
	"github.com/lariv-in/p_filesystem"
	"github.com/lariv-in/p_users"
	"gorm.io/gorm"
)

type Student struct {
	gorm.Model

	UserID    uint         `gorm:"uniqueIndex;notnull"`
	User      p_users.User `gorm:"constraint:OnDelete:CASCADE"`
	StudentNo string       `gorm:"uniqueIndex;notnull"`
	DOB       *time.Time

	Assets []p_filesystem.VNode `gorm:"many2many:student_assets;"`
}

func init() {
	lago.OnDBInit(func(d *gorm.DB) *gorm.DB {
		if err := d.AutoMigrate(&Student{}); err != nil {
			log.Panicf("failed to migrate Student model: %v", err)
		}
		d.FirstOrCreate(&p_users.Role{}, p_users.Role{Name: "student"})
		return d
	})

	lago.RegistryAdmin.Register("p_students", lago.AdminPanel[Student]{
		SearchField: "StudentNo",
		ListFields:  []string{"StudentNo", "User.Name", "DOB", "UpdatedAt"},
		Preload:     []string{"User"},
	})
}
