package p_students

import (
	"fmt"
	"log/slog"
	"math/rand"
	"time"

	"github.com/lariv-in/lago"
	"github.com/lariv-in/p_users"
	"gorm.io/gorm"
)

const defaultPassword = "Pass1234#"

func generateStudentNo(index int) string {
	return fmt.Sprintf("STU%05d", index+1)
}

func randomDOB() *time.Time {
	// ~25% chance of nil DOB
	if rand.Intn(100) < 25 {
		return nil
	}
	// Random DOB between ages 5 and 20 (school-age range)
	now := time.Now()
	yearsAgo := rand.Intn(16) + 5 // 5 to 20
	daysOffset := rand.Intn(365)
	dob := time.Date(now.Year()-yearsAgo, 1, 1+daysOffset, 0, 0, 0, 0, time.UTC)
	return &dob
}

// CreateSampleStudent idempotently creates a sample student (student1@lariv.in)
// with a known password for development/testing. Returns the existing student
// if already present.
func CreateSampleStudent(db *gorm.DB) (*Student, error) {
	const sampleEmail = "student1@lariv.in"

	var existing Student
	err := db.Joins("User").Where("\"User\".email = ?", sampleEmail).First(&existing).Error
	if err == nil {
		fmt.Println("Sample student (student1) already exists")
		return &existing, nil
	}

	role := p_users.Role{Name: "student"}
	db.Where("name = ?", "student").FirstOrCreate(&role)

	user := p_users.User{
		Name:     "Sample Student",
		Email:    sampleEmail,
		Phone:    p_users.GenerateRandomPhone(),
		Password: []byte(defaultPassword),
		RoleID:   role.ID,
	}
	if err := db.Create(&user).Error; err != nil {
		return nil, fmt.Errorf("failed to create sample student user: %w", err)
	}

	student := Student{
		UserID:    user.ID,
		StudentNo: "STU00000",
		DOB:       nil,
	}
	if err := db.Create(&student).Error; err != nil {
		return nil, fmt.Errorf("failed to create sample student: %w", err)
	}

	fmt.Println("Created sample student (student1@lariv.in)")
	return &student, nil
}

func init() {
	lago.RegistryGenerator.Register("students.Generator", lago.Generator{
		Create: func(db *gorm.DB) error {
			// Create the sample student first
			if _, err := CreateSampleStudent(db); err != nil {
				return err
			}

			const studentCount = 30

			for i := range studentCount {
				user, err := p_users.GenerateUserWithoutPassword(db, "student")
				if err != nil {
					return fmt.Errorf("failed to generate user for student %d: %w", i, err)
				}

				studentNo := generateStudentNo(i)
				dob := randomDOB()

				student := Student{
					UserID:    user.ID,
					StudentNo: studentNo,
					DOB:       dob,
				}
				if err := db.Create(&student).Error; err != nil {
					return fmt.Errorf("failed to create student %s: %w", studentNo, err)
				}
			}

			fmt.Printf("Created %d students (+ 1 sample)\n", studentCount)
			return nil
		},
		Remove: func(db *gorm.DB) error {
			// Clear many-to-many join table before deleting students
			if err := db.Exec("DELETE FROM student_assets").Error; err != nil {
				slog.Error("failed clearing student_assets join table", "error", err)
			}
			return db.Unscoped().Where("1=1").Delete(&Student{}).Error
		},
	})
}
