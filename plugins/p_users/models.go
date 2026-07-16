package p_users

import (
	"crypto/rand"
	"log"

	"github.com/lariv-in/lariv"
	"github.com/lariv-in/lariv/registry"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model

	Name string `gorm:"notnull"`
	// Use `unique` not `uniqueIndex`: otherwise PostgreSQL + GORM AutoMigrate
	// tries DROP CONSTRAINT uni_* on an index-backed uniqueness (42704).
	Email        string `gorm:"unique"`
	Phone        string `gorm:"unique"`
	IsSuperuser  bool   `gorm:"notnull"`
	RoleID       uint   `gorm:"notnull;default:1"`
	Role         Role   `gorm:"notnull"`
	Password     []byte `gorm:"-"`
	PasswordHash []byte `gorm:"column:password"`
	PasswordSalt []byte
	Timezone     string `gorm:"default:Asia/Kolkata"`
}

type Role struct {
	gorm.Model
	Name string `gorm:"unique"`
}

func (u *User) BeforeSave(tx *gorm.DB) (err error) {
	if len(u.Password) > 0 {
		return u.hashPassword()
	}
	return nil
}

func (u *User) hashPassword() (err error) {
	u.PasswordSalt = make([]byte, 256)
	// Never actually errors out and always fills the buffer
	n, err := rand.Read(u.PasswordSalt)
	if err != nil {
		log.Panicf("This should never happen, crypto read err while hashing user password: %e", err)
	}

	if n != 256 {
		log.Panicf("This should never happen, password salt n = %d", n)
	}

	u.PasswordHash = HashPassword(u.Password, u.PasswordSalt)
	u.Password = nil
	return err
}

func pluginModels() lariv.PluginFeatures[any] {
	return lariv.PluginFeatures[any]{
		Entries: []registry.Pair[string, any]{
			{Key: "p_users.Role", Value: Role{}},
			{Key: "p_users.User", Value: User{}},
		},
	}
}

func pluginDBInitHooks() lariv.PluginFeatures[lariv.DBInitHook] {
	return lariv.PluginFeatures[lariv.DBInitHook]{
		Entries: []registry.Pair[string, lariv.DBInitHook]{{
			Key: "p_users.bootstrap",
			Value: func(d *gorm.DB) *gorm.DB {
				// Ensure ID 1 is always the safe "Unassigned" fallback role (Attrs applies on insert only).
				var unassigned Role
				d.Attrs(Role{Model: gorm.Model{ID: 1}}).FirstOrCreate(&unassigned, Role{Name: "unassigned"})
				return d
			},
		}},
	}
}

func init() {
	lariv.RegistryAdmin.Register("p_users", lariv.AdminPanel[User]{
		SearchField: "Name",
		ListFields:  []string{"Name", "Email", "IsSuperuser", "Role.Name"},
		Preload:     []string{"Role"},
	})
}
