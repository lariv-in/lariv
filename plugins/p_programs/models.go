package p_programs

import (
	"github.com/lariv-in/lago/lago"
	"gorm.io/gorm"
)

// Program represents the base program entity.
// Django model: name, code (unique), description.
type Program struct {
	gorm.Model

	Name        string
	Code        string `gorm:"uniqueIndex"`
	Description string
}

func init() {
	lago.OnDBInit(func(d *gorm.DB) *gorm.DB {
		d.AutoMigrate(Program{})
		return d
	})

	lago.RegistryAdmin.Register("p_programs", lago.AdminPanel[Program]{
		SearchField: "Name",
		ListFields:  []string{"Name", "Code"},
	})
}

