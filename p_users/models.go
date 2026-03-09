package p_users

import (
	"crypto/rand"
	"fmt"

	"github.com/lariv-in/lago"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model

	Name         string `gorm:"notnull"`
	Email        string `gorm:"uniqueIndex"`
	Phone        string `gorm:"uniqueIndex"`
	IsSuperuser  bool   `gorm:"notnull"`
	RoleID       int    `gorm:"notnull"`
	Role         Role   `gorm:"notnull"`
	Password     []byte `gorm:"notnull"`
	PasswordSalt []byte `gorm:"notnull"`
}

type Role struct {
	gorm.Model
	Name string `gorm:"unique"`
}

func (u *User) BeforeSave(tx *gorm.DB) (err error) {
	if len(u.Password) != 0 {
		u.PasswordSalt = make([]byte, 256)

		// Never actually errors out and always fills the buffer
		_, _ = rand.Read(u.PasswordSalt)
		u.Password = HashPassword(u.Password, u.PasswordSalt)
	}
	fmt.Println(*u)
	return nil
}

func init() {
	lago.OnDbInit(func(d *gorm.DB) *gorm.DB {
		d.AutoMigrate(User{})
		d.AutoMigrate(Role{})
		return d
	})
}
