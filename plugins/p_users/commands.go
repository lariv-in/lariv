package p_users

import (
	"context"
	"fmt"
	"log/slog"
	"net/mail"

	"github.com/lariv-in/lago/lago"
	"github.com/nyaruka/phonenumbers"
	"github.com/spf13/cobra"
	"gorm.io/gorm"
)

func init() {
	lago.RegistryCommand.Register("p_users.createsuperuser", createSuperuserCommand)
	lago.RegistryCommand.Register("p_users.changepassword", changePasswordCommand)
	lago.RegistryCommand.Register("p_users.revalidate_users", revalidateUsersCommand)
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

			role := Role{Name: "unassigned"}
			if err := db.Where("name = ?", role.Name).FirstOrCreate(&role).Error; err != nil {
				return fmt.Errorf("failed to fetch or create Unassigned role: %w", err)
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
			if err := gorm.G[User](db).Create(context.Background(), &user); err != nil {
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
			user, err := gorm.G[User](db).Where("email = ?", email).First(context.Background())
			if err != nil {
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

func revalidateUsersCommand(config lago.LagoConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "revalidate_users",
		Short: "Re-parse and normalize all user email and phone fields",
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := lago.InitDB(config)
			if err != nil {
				return err
			}

			users, err := gorm.G[User](db).Find(context.Background())
			if err != nil {
				return fmt.Errorf("failed to fetch users for revalidation: %w", err)
			}

			var (
				total   = len(users)
				updated int
				skipped int
			)

			for _, user := range users {
				originalEmail := user.Email
				originalPhone := user.Phone

				// Parse and normalize email
				if originalEmail != "" {
					addr, err := mail.ParseAddress(originalEmail)
					if err != nil {
						slog.Warn("Failed to parse user email during revalidation",
							"user_id", user.ID,
							"email", originalEmail,
							"name", user.Name,
							"err", err,
						)
						skipped++
						continue
					}
					user.Email = addr.Address
				}

				// Parse and normalize phone
				if originalPhone != "" {
					num, err := phonenumbers.Parse(originalPhone, "IN")
					if err != nil {
						slog.Warn("Failed to parse user phone during revalidation",
							"user_id", user.ID,
							"phone", originalPhone,
							"name", user.Name,
							"err", err,
						)
						skipped++
						continue
					}
					user.Phone = phonenumbers.Format(num, phonenumbers.E164)
				}

				if err := db.Save(&user).Error; err != nil {
					slog.Warn("Failed to save user during revalidation",
						"user_id", user.ID,
						"name", user.Name,
						"err", err,
					)
					skipped++
					continue
				}

				updated++
			}

			fmt.Printf("Revalidation complete. Total users: %d, updated: %d, skipped: %d\n", total, updated, skipped)
			return nil
		},
	}

	return cmd
}
