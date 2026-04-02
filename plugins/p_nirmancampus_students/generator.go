package p_nirmancampus_students

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"time"

	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
	"gorm.io/gorm"
)

const defaultPassword = "Pass1234#"

func generateStudentNo(index int) string {
	return fmt.Sprintf("STU%05d", index+1)
}

func randomDOB() *time.Time {
	if rand.Intn(100) < 25 {
		return nil
	}
	now := time.Now()
	yearsAgo := rand.Intn(16) + 5
	daysOffset := rand.Intn(365)
	dob := time.Date(now.Year()-yearsAgo, 1, 1+daysOffset, 0, 0, 0, 0, time.UTC)
	return &dob
}

var studentCategories = []string{
	"General",
	"OBC",
	"SC",
	"ST",
	"",
}

var fathersNamePrefixes = []string{
	"Ravi",
	"Suresh",
	"Mahesh",
	"Ramesh",
	"Venkatesh",
	"Kumar",
	"Prakash",
	"Rajesh",
	"Mohan",
	"Raghav",
}

func randomAddress(r *rand.Rand) string {
	number := r.Intn(9999) + 1
	street := []string{"Main St", "Lake View", "Market Rd", "Park Ave", "Temple Rd"}[r.Intn(5)]
	city := []string{"Nirmancampus", "Hyderabad", "Pune", "Chennai", "Delhi"}[r.Intn(5)]
	pin := r.Intn(899999) + 100000
	return fmt.Sprintf("%s %d, %s - %d", street, number, city, pin)
}

func randomFathersName(r *rand.Rand) string {
	if r.Intn(100) < 30 {
		return ""
	}
	prefix := fathersNamePrefixes[r.Intn(len(fathersNamePrefixes))]
	suffix := r.Intn(999) + 1
	return fmt.Sprintf("%s %d", prefix, suffix)
}

func randomNirmancampusFields(r *rand.Rand) (fathersName, category, address string) {
	fathersName = randomFathersName(r)
	category = studentCategories[r.Intn(len(studentCategories))]
	if r.Intn(100) < 60 {
		address = randomAddress(r)
	}
	return fathersName, category, address
}

// CreateSampleStudent idempotently creates a sample student (student1@lariv.in).
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
	if err := gorm.G[p_users.User](db).Create(context.Background(), &user); err != nil {
		return nil, fmt.Errorf("failed to create sample student user: %w", err)
	}

	student := Student{
		UserID:      user.ID,
		StudentNo:   "STU00000",
		DOB:         nil,
		FathersName: "",
		Category:    "",
		Address:     "",
	}
	if err := gorm.G[Student](db).Create(context.Background(), &student); err != nil {
		return nil, fmt.Errorf("failed to create sample student: %w", err)
	}

	fmt.Println("Created sample student (student1@lariv.in)")
	return &student, nil
}

func init() {
	lago.RegistryGenerator.Register("students.Generator", lago.Generator{
		Create: func(db *gorm.DB) error {
			if _, err := CreateSampleStudent(db); err != nil {
				return err
			}

			const studentCount = 30
			r := rand.New(rand.NewSource(time.Now().UnixNano()))

			for i := range studentCount {
				user, err := p_users.GenerateUserWithoutPassword(db, "student")
				if err != nil {
					return fmt.Errorf("failed to generate user for student %d: %w", i, err)
				}

				studentNo := generateStudentNo(i)
				dob := randomDOB()
				fn, cat, addr := randomNirmancampusFields(r)

				student := Student{
					UserID:      user.ID,
					StudentNo:   studentNo,
					DOB:         dob,
					FathersName: fn,
					Category:    cat,
					Address:     addr,
				}
				if err := gorm.G[Student](db).Create(context.Background(), &student); err != nil {
					return fmt.Errorf("failed to create student %s: %w", studentNo, err)
				}
			}

			fmt.Printf("Created %d students (+ 1 sample)\n", studentCount)
			return nil
		},
		Remove: func(db *gorm.DB) error {
			if err := db.Exec("DELETE FROM student_assets").Error; err != nil {
				slog.Error("failed clearing student_assets join table", "error", err)
			}
			return db.Unscoped().Where("1=1").Delete(&Student{}).Error
		},
	})
}
