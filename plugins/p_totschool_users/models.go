package p_totschool_users

import (
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/p_users"
	"gorm.io/gorm"
)

func init() {
	lago.OnDBInit(func(d *gorm.DB) *gorm.DB {
		d.FirstOrCreate(&p_users.Role{}, p_users.Role{Name: "totschool_student"})
		d.FirstOrCreate(&p_users.Role{}, p_users.Role{Name: "totschool_admin"})
		return d
	})
}
