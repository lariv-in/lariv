package p_users

import (
	"crypto/rand"
	"log"

	"github.com/lariv-in/lago"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model

	Name         string `gorm:"notnull"`
	Email        string `gorm:"uniqueIndex"`
	Phone        string `gorm:"uniqueIndex"`
	IsSuperuser  bool   `gorm:"notnull"`
	RoleID       uint    `gorm:"notnull"`
	Role         Role   `gorm:"notnull"`
	Password     []byte `gorm:"notnull"`
	PasswordSalt []byte `gorm:"notnull"`
	Timezone     string `gorm:"default:Asia/Kolkata"`
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
	return
}

func (u *User) hashPassword() (err error) {
	if len(u.Password) == 0 {
		return
	}
	u.PasswordSalt = make([]byte, 256)
	// Never actually errors out and always fills the buffer
	n, err := rand.Read(u.PasswordSalt)
	if err != nil {
		log.Panicf("This should never happen, crypto read err while hashing user password: %e", err)
	}

	if n != 256 {
		log.Panicf("This should never happen, password salt n = %d", n)
	}

	u.Password = HashPassword(u.Password, u.PasswordSalt)
	return
}

func init() {
	lago.OnDBInit(func(d *gorm.DB) *gorm.DB {
		d.AutoMigrate(User{})
		d.AutoMigrate(Role{})
		return d
	})
	lago.RegistryAdmin.Register("p_users", lago.AdminPanel[User]{
		SearchField: "Name",
		ListFields:  []string{"Name", "Email", "IsSuperuser", "Role.Name"},
		Preload:     []string{"Role"},
	})
}
