package p_users

import (
	"fmt"

	"github.com/lariv-in/lago"
	"github.com/spf13/cobra"
)

func init() {
	lago.RegistryCommand.Register("p_users.createsuperuser", createSuperuserCommand)
	lago.RegistryCommand.Register("p_users.changepassword", changePasswordCommand)
}

func createSuperuserCommand(config lago.LagoConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "createsuperuser",
		Short: "Create a superuser account",
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := lago.InitDB(config)
			if err != nil {
				return err
			}

			name, _ := cmd.Flags().GetString("name")
			email, _ := cmd.Flags().GetString("email")
			phone, _ := cmd.Flags().GetString("phone")
			password, _ := cmd.Flags().GetString("password")

			var role Role
			if err := db.Model(Role{}).Order("id ASC").First(&role).Error; err != nil {
				return fmt.Errorf("Couldn't fetch roles, err: %v", err)
			}

			user := User{
				Name:        name,
				Email:       email,
				Phone:       phone,
				Password:    []byte(password),
				IsSuperuser: true,
				RoleID:      role.ID,
				Role:        role,
			}
			if err := db.Create(&user).Error; err != nil {
				return fmt.Errorf("failed to create superuser: %w", err)
			}

			fmt.Printf("Superuser %q (%s) created successfully.\n", name, email)
			return nil
		},
	}

	cmd.Flags().String("name", "", "Name of the superuser")
	cmd.Flags().String("email", "", "Email of the superuser")
	cmd.Flags().String("phone", "", "Phone number of the superuser")
	cmd.Flags().String("password", "", "Password for the superuser")
	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("email")
	cmd.MarkFlagRequired("phone")
	cmd.MarkFlagRequired("password")

	return cmd
}

func changePasswordCommand(config lago.LagoConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "changepassword",
		Short: "Change a user's password by email",
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := lago.InitDB(config)
			if err != nil {
				return err
			}

			email, _ := cmd.Flags().GetString("email")
			password, _ := cmd.Flags().GetString("password")
			var user User
			if err := db.Where("email = ?", email).First(&user).Error; err != nil {
				return fmt.Errorf("failed to find user: %w", err)
			}

			user.Password = []byte(password)

			if err := db.Save(&user).Error; err != nil {
				return fmt.Errorf("failed to update password: %w", err)
			}

			fmt.Println("Password updated, user", user)
			return nil
		},
	}

	cmd.Flags().String("email", "", "Email of the user")
	cmd.Flags().String("password", "", "New password")
	cmd.MarkFlagRequired("email")
	cmd.MarkFlagRequired("password")

	return cmd
}
