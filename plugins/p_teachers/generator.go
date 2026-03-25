package p_teachers

import (
	"fmt"
	"log/slog"
	"math/rand"

	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
	"gorm.io/gorm"
)

const defaultPassword = "Pass1234#"

var qualificationsList = []string{
	"B.Ed, M.A. English",
	"M.Sc. Mathematics, B.Ed",
	"Ph.D. Physics",
	"M.A. Hindi, B.Ed",
	"M.Sc. Chemistry, B.Ed",
	"M.A. History, B.Ed",
	"M.Sc. Biology, NET Qualified",
	"M.Com, B.Ed",
	"M.A. Political Science, B.Ed",
	"M.Sc. Computer Science",
	"M.A. Economics, Ph.D.",
	"B.Tech, B.Ed",
	"M.A. Sanskrit, B.Ed",
	"M.Sc. Environmental Science",
	"M.F.A., B.Ed",
	"M.P.Ed (Physical Education)",
	"M.A. Geography, B.Ed",
	"M.Sc. Statistics, B.Ed",
	"M.A. Sociology, NET Qualified",
	"M.Mus (Music), B.Ed",
}

func generateTeacherCode(index int) string {
	return fmt.Sprintf("TCH%03d", index+1)
}

func randomQualifications() *string {
	// ~20% chance of nil qualifications
	if rand.Intn(100) < 20 {
		return nil
	}
	q := qualificationsList[rand.Intn(len(qualificationsList))]
	return &q
}

// CreateSampleTeacher idempotently creates a sample teacher (teacher1@lariv.in)
// with a known password for development/testing. Returns the existing teacher
// if already present.
func CreateSampleTeacher(db *gorm.DB) (*Teacher, error) {
	const sampleEmail = "teacher1@lariv.in"

	var existing Teacher
	err := db.Joins("User").Where("\"User\".email = ?", sampleEmail).First(&existing).Error
	if err == nil {
		fmt.Println("Sample teacher (teacher1) already exists")
		return &existing, nil
	}

	role := p_users.Role{Name: "teacher"}
	db.Where("name = ?", "teacher").FirstOrCreate(&role)

	qualifications := "M.Sc. Computer Science"

	user := p_users.User{
		Name:     "Sample Teacher",
		Email:    sampleEmail,
		Phone:    p_users.GenerateRandomPhone(),
		Password: []byte(defaultPassword),
		RoleID:   role.ID,
	}
	if err := db.Create(&user).Error; err != nil {
		return nil, fmt.Errorf("failed to create sample teacher user: %w", err)
	}

	teacher := Teacher{
		UserID:         user.ID,
		Code:           "TCH000",
		Qualifications: &qualifications,
	}
	if err := db.Create(&teacher).Error; err != nil {
		return nil, fmt.Errorf("failed to create sample teacher: %w", err)
	}

	fmt.Println("Created sample teacher (teacher1@lariv.in)")
	return &teacher, nil
}

func init() {
	lago.RegistryGenerator.Register("teachers.Generator", lago.Generator{
		Create: func(db *gorm.DB) error {
			// Create the sample teacher first
			if _, err := CreateSampleTeacher(db); err != nil {
				return err
			}

			const teacherCount = 20

			for i := range teacherCount {
				user, err := p_users.GenerateUserWithoutPassword(db, "teacher")
				if err != nil {
					return fmt.Errorf("failed to generate user for teacher %d: %w", i, err)
				}

				code := generateTeacherCode(i)
				qualifications := randomQualifications()

				teacher := Teacher{
					UserID:         user.ID,
					Code:           code,
					Qualifications: qualifications,
				}
				if err := db.Create(&teacher).Error; err != nil {
					return fmt.Errorf("failed to create teacher %s: %w", code, err)
				}
			}

			fmt.Printf("Created %d teachers\n", teacherCount)
			return nil
		},
		Remove: func(db *gorm.DB) error {
			// Clear many-to-many join table before deleting teachers
			if err := db.Exec("DELETE FROM teacher_assets").Error; err != nil {
				slog.Error("failed clearing teacher_assets join table", "error", err)
			}
			return db.Unscoped().Where("1=1").Delete(&Teacher{}).Error
		},
	})
}
