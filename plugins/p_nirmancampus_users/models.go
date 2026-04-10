package p_nirmancampus_users

import (
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
	"gorm.io/gorm"
)

func init() {
	lago.OnDBInit("p_nirmancampus_users.admin_role", func(d *gorm.DB) *gorm.DB {
		d.FirstOrCreate(&p_users.Role{}, p_users.Role{Name: "admin"})
		return d
	})
}
