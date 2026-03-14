package p_users

import (
	"crypto/rand"

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

func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	return u.hashPassword()
}

func (u *User) BeforeUpdate(tx *gorm.DB) (err error) {
	if tx.Statement.Changed("Password") {
		return u.hashPassword()
	}
	return nil
}

func (u *User) hashPassword() error {
	if len(u.Password) == 0 {
		return nil
	}
	u.PasswordSalt = make([]byte, 256)
	// Never actually errors out and always fills the buffer
	_, _ = rand.Read(u.PasswordSalt)
	u.Password = HashPassword(u.Password, u.PasswordSalt)
	return nil
}

func init() {
	lago.OnDBInit(func(d *gorm.DB) *gorm.DB {
		d.AutoMigrate(User{})
		d.AutoMigrate(Role{})
		return d
	})
	lago.RegistryAdmin.Register("p_users", lago.AdminPanel[User]{SearchField: "Name", ListFields: []string{"Name", "Email", "IsSuperuser", "Role.Name"}, Preload: []string{"Role"}})
}
