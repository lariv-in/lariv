package p_teachers

import (
	"log"

	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/p_filesystem"
	"github.com/lariv-in/lago/p_users"
	"gorm.io/gorm"
)

type Teacher struct {
	gorm.Model

	UserID         uint                `gorm:"uniqueIndex;notnull"`
	User           p_users.User        `gorm:"constraint:OnDelete:CASCADE"`
	ProfilePhotoID *uint               `gorm:"index"`
	ProfilePhoto   *p_filesystem.VNode `gorm:"constraint:OnDelete:SET NULL"`
	Code           string              `gorm:"uniqueIndex;notnull"`
	Qualifications *string
	Assets         []p_filesystem.VNode `gorm:"many2many:teacher_assets;"`
}

func init() {
	lago.OnDBInit(func(d *gorm.DB) *gorm.DB {
		if err := d.AutoMigrate(&Teacher{}); err != nil {
			log.Panicf("failed to migrate Teacher model: %v", err)
		}
		d.FirstOrCreate(&p_users.Role{}, p_users.Role{Name: "teacher"})
		return d
	})

	lago.RegistryAdmin.Register("p_teachers", lago.AdminPanel[Teacher]{
		SearchField: "Code",
		ListFields:  []string{"Code", "User.Name", "Qualifications", "UpdatedAt"},
		Preload:     []string{"User"},
	})
}
