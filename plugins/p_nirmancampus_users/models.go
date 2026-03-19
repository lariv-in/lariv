package p_nirmancampus_users

import (
	"github.com/lariv-in/lago"
	"github.com/lariv-in/p_users"
	"gorm.io/gorm"
)

func init() {
	lago.OnDBInit(func(d *gorm.DB) *gorm.DB {
		d.FirstOrCreate(&p_users.Role{}, p_users.Role{Name: "nirmancampus_admin"})
		d.FirstOrCreate(&p_users.Role{}, p_users.Role{Name: "nirmancampus_teacher"})
		d.FirstOrCreate(&p_users.Role{}, p_users.Role{Name: "nirmancampus_student"})
		return d
	})
}
