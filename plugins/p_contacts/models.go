package p_contacts

import (
	"github.com/lariv-in/lago/lago"
	"gorm.io/gorm"
)

// Contact is a stored person or organization contact.
type Contact struct {
	gorm.Model

	Name    string `gorm:"notnull"`
	Phone   string
	Email   string
	Address string `gorm:"type:text"`
	Notes   string `gorm:"type:text"`
}

func init() {
	lago.OnDBInit(func(d *gorm.DB) *gorm.DB {
		lago.RegisterModel[Contact](d)
		return d
	})

	lago.RegistryAdmin.Register("p_contacts", lago.AdminPanel[Contact]{
		SearchField: "Name",
		ListFields: []string{
			"Name",
			"Phone",
			"Email",
			"Address",
			"Notes",
			"CreatedAt",
			"UpdatedAt",
		},
	})
}
