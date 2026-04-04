package p_nirmancampus_users

import (
	"context"
	"fmt"

	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
	"gorm.io/gorm"
)

const defaultPassword = "Pass1234#"

// CreateSampleAdmin idempotently creates a sample admin user (admin1@lariv.in).
func CreateSampleAdmin(db *gorm.DB) (*p_users.User, error) {
	const sampleEmail = "admin1@lariv.in"

	existing, err := gorm.G[p_users.User](db).Where("email = ?", sampleEmail).First(context.Background())
	if err == nil {
		fmt.Println("Sample admin (admin1) already exists")
		return &existing, nil
	}

	role := p_users.Role{Name: "admin"}
	db.Where("name = ?", "admin").FirstOrCreate(&role)

	user := p_users.User{
		Name:     "Sample Admin",
		Email:    sampleEmail,
		Phone:    p_users.GenerateRandomPhone(),
		Password: []byte(defaultPassword),
		RoleID:   role.ID,
	}
	if err := gorm.G[p_users.User](db).Create(context.Background(), &user); err != nil {
		return nil, fmt.Errorf("failed to create sample admin user: %w", err)
	}

	fmt.Println("Created sample admin (admin1@lariv.in)")
	return &user, nil
}

func init() {
	lago.RegistryGenerator.Register("nirmancampus_users.Generator", lago.Generator{
		Create: func(db *gorm.DB) error {
			_, err := CreateSampleAdmin(db)
			return err
		},
		Remove: func(db *gorm.DB) error {
			return db.Unscoped().Where("email = ?", "admin1@lariv.in").Delete(&p_users.User{}).Error
		},
	})
}
