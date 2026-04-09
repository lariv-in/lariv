package p_nirmancampus_students

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"strings"
	"time"

	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/registry"
	"github.com/lariv-in/lago/plugins/p_filesystem"
	"github.com/lariv-in/lago/plugins/p_users"
	"gorm.io/gorm"
)

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

var mothersNamePrefixes = []string{
	"Sunita",
	"Priya",
	"Kavita",
	"Anita",
	"Meera",
	"Lakshmi",
	"Radha",
	"Geeta",
}

func randomMothersName(r *rand.Rand) string {
	if r.Intn(100) < 30 {
		return ""
	}
	prefix := mothersNamePrefixes[r.Intn(len(mothersNamePrefixes))]
	suffix := r.Intn(999) + 1
	return fmt.Sprintf("%s %d", prefix, suffix)
}

func randomFathersName(r *rand.Rand) string {
	if r.Intn(100) < 30 {
		return ""
	}
	prefix := fathersNamePrefixes[r.Intn(len(fathersNamePrefixes))]
	suffix := r.Intn(999) + 1
	return fmt.Sprintf("%s %d", prefix, suffix)
}

// studentCategoryKeys lists persisted category keys in the same order as StudentCategoryChoices.
var studentCategoryKeys = registry.KeysFromPairs(StudentCategoryChoices)

func randomNirmancampusFields(r *rand.Rand) (motherName, fatherName, category, address string) {
	motherName = randomMothersName(r)
	fatherName = randomFathersName(r)
	category = studentCategoryKeys[r.Intn(len(studentCategoryKeys))]
	if r.Intn(100) < 60 {
		address = randomAddress(r)
	}
	return motherName, fatherName, category, address
}

func pickDistinctFiles(files []p_filesystem.VNode, n int, excludeID *uint) []p_filesystem.VNode {
	if n <= 0 || len(files) == 0 {
		return nil
	}
	order := rand.Perm(len(files))
	var out []p_filesystem.VNode
	for _, i := range order {
		f := files[i]
		if excludeID != nil && f.ID == *excludeID {
			continue
		}
		out = append(out, f)
		if len(out) >= n {
			break
		}
	}
	return out
}

func loadFileNodes(db *gorm.DB) ([]p_filesystem.VNode, error) {
	return gorm.G[p_filesystem.VNode](db).Where("is_directory = ?", false).Find(context.Background())
}

// assignStudentPhoto picks an existing file (usually) or generates one; updates files when a new file is created.
func assignStudentPhoto(db *gorm.DB, files []p_filesystem.VNode) (photoID *uint, filesOut []p_filesystem.VNode, err error) {
	filesOut = files
	if len(filesOut) > 0 && rand.Intn(100) < 80 {
		picked := filesOut[rand.Intn(len(filesOut))]
		id := picked.ID
		return &id, filesOut, nil
	}
	node, genErr := p_filesystem.GeneratePhotoFile(db)
	if genErr != nil {
		return nil, filesOut, genErr
	}
	if node != nil {
		id := node.ID
		filesOut, err = loadFileNodes(db)
		if err != nil {
			return nil, filesOut, err
		}
		return &id, filesOut, nil
	}
	return nil, filesOut, nil
}

func attachRandomStudentDocuments(db *gorm.DB, student *Student) error {
	nDocs := rand.Intn(4)
	if nDocs == 0 {
		return nil
	}
	files, err := loadFileNodes(db)
	if err != nil {
		return err
	}
	docs := pickDistinctFiles(files, nDocs, student.PhotoID)
	if len(docs) == 0 {
		return nil
	}
	return db.Model(student).Association("Documents").Append(docs)
}

func randomStudentContact(db *gorm.DB, r *rand.Rand) (name, email, phone string, err error) {
	name = p_users.GetRandomIndianName()
	studentCount, err := gorm.G[Student](db).Count(context.Background(), "*")
	if err != nil {
		return "", "", "", err
	}
	username := fmt.Sprintf("%s_%d", strings.ToLower(strings.ReplaceAll(name, " ", ".")), studentCount+1)
	email = fmt.Sprintf("%s@school1.com", username)
	phone = p_users.GenerateRandomPhone()
	return name, email, phone, nil
}

// CreateSampleStudent idempotently creates a sample student (student1@lariv.in).
func CreateSampleStudent(db *gorm.DB) (*Student, error) {
	const sampleEmail = "student1@lariv.in"

	var existing Student
	err := db.Where("email = ?", sampleEmail).First(&existing).Error
	if err == nil {
		fmt.Println("Sample student (student1) already exists")
		return &existing, nil
	}

	dob := time.Date(2010, 6, 15, 0, 0, 0, 0, time.UTC)
	student := Student{
		Name:       "Sample Student",
		Email:      sampleEmail,
		Phone:      p_users.GenerateRandomPhone(),
		StudentNo:  "STU00000",
		DOB:        &dob,
		MotherName: "",
		FatherName: "",
		Category:   "",
		Address:    "",
	}
	files, err := loadFileNodes(db)
	if err != nil {
		return nil, fmt.Errorf("load filesystem files for sample student: %w", err)
	}
	photoID, _, err := assignStudentPhoto(db, files)
	if err != nil {
		return nil, fmt.Errorf("generate photo for sample student: %w", err)
	}
	student.PhotoID = photoID
	if err := gorm.G[Student](db).Create(context.Background(), &student); err != nil {
		return nil, fmt.Errorf("failed to create sample student: %w", err)
	}
	if err := attachRandomStudentDocuments(db, &student); err != nil {
		return nil, fmt.Errorf("attach documents for sample student: %w", err)
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

			files, err := loadFileNodes(db)
			if err != nil {
				return err
			}

			for i := range studentCount {
				name, email, phone, err := randomStudentContact(db, r)
				if err != nil {
					return fmt.Errorf("failed to generate contact for student %d: %w", i, err)
				}

				studentNo := generateStudentNo(i)
				dob := randomDOB()
				mn, fn, cat, addr := randomNirmancampusFields(r)

				student := Student{
					Name:       name,
					Email:      email,
					Phone:      phone,
					StudentNo:  studentNo,
					DOB:        dob,
					MotherName: mn,
					FatherName: fn,
					Category:   cat,
					Address:    addr,
				}
				var photoErr error
				student.PhotoID, files, photoErr = assignStudentPhoto(db, files)
				if photoErr != nil {
					return fmt.Errorf("generate photo for student %s: %w", studentNo, photoErr)
				}
				if err := gorm.G[Student](db).Create(context.Background(), &student); err != nil {
					return fmt.Errorf("failed to create student %s: %w", studentNo, err)
				}
				if err := attachRandomStudentDocuments(db, &student); err != nil {
					return fmt.Errorf("attach documents for student %s: %w", studentNo, err)
				}
			}

			fmt.Printf("Created %d students (+ 1 sample)\n", studentCount)
			return nil
		},
		Remove: func(db *gorm.DB) error {
			if err := db.Exec("DELETE FROM student_documents").Error; err != nil {
				slog.Error("failed clearing student_documents join table", "error", err)
			}
			return db.Unscoped().Where("1=1").Delete(&Student{}).Error
		},
	})
}
