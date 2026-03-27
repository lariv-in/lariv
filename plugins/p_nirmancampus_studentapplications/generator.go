package p_nirmancampus_studentapplications

import (
	"fmt"
	"log/slog"
	"math/rand"
	"time"

	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_filesystem"
	"github.com/lariv-in/lago/plugins/p_nirmancampus_programs"
	"github.com/lariv-in/lago/plugins/p_users"
	"gorm.io/gorm"
)

// sampleApplicationRows are seed records; ProgramID is chosen from existing programs at runtime.
var sampleApplicationRows = []struct {
	studentName     string
	email           string
	dob             time.Time
	motherName      string
	fatherName      string
	category        string
	completeAddress string
	mobile          string
}{
	{
		studentName:     "Riya Sharma",
		email:           "riya.sharma@example.edu",
		dob:             time.Date(2005, 3, 12, 0, 0, 0, 0, time.UTC),
		motherName:      "Sunita Sharma",
		fatherName:      "Rajesh Sharma",
		category:        "General",
		completeAddress: "42, Station Road, Indore, MP 452001",
		mobile:          "9876501234",
	},
	{
		studentName:     "Arjun Mehta",
		email:           "arjun.mehta@example.edu",
		dob:             time.Date(2004, 7, 22, 0, 0, 0, 0, time.UTC),
		motherName:      "Neha Mehta",
		fatherName:      "Vikram Mehta",
		category:        "OBC",
		completeAddress: "18, Lake View Colony, Ujjain, MP 456010",
		mobile:          "9876502234",
	},
	{
		studentName:     "Ananya Iyer",
		email:           "ananya.iyer@example.edu",
		dob:             time.Date(2006, 1, 5, 0, 0, 0, 0, time.UTC),
		motherName:      "Lakshmi Iyer",
		fatherName:      "Karthik Iyer",
		category:        "General",
		completeAddress: "7, Teachers Colony, Bhopal, MP 462003",
		mobile:          "9876503234",
	},
	{
		studentName:     "Mohammed Khan",
		email:           "mohammed.khan@example.edu",
		dob:             time.Date(2005, 11, 18, 0, 0, 0, 0, time.UTC),
		motherName:      "Ayesha Khan",
		fatherName:      "Salim Khan",
		category:        "General",
		completeAddress: "91, Old City, Burhanpur, MP 450331",
		mobile:          "9876504234",
	},
	{
		studentName:     "Priya Nair",
		email:           "priya.nair@example.edu",
		dob:             time.Date(2004, 9, 30, 0, 0, 0, 0, time.UTC),
		motherName:      "Deepa Nair",
		fatherName:      "Suresh Nair",
		category:        "SC",
		completeAddress: "3B, Riverside Apartments, Jabalpur, MP 482001",
		mobile:          "9876505234",
	},
	{
		studentName:     "Kavya Reddy",
		email:           "kavya.reddy@example.edu",
		dob:             time.Date(2005, 4, 8, 0, 0, 0, 0, time.UTC),
		motherName:      "Padma Reddy",
		fatherName:      "Srinivas Reddy",
		category:        "General",
		completeAddress: "55, MG Road, Gwalior, MP 474001",
		mobile:          "9876506234",
	},
	{
		studentName:     "Dev Patel",
		email:           "dev.patel@example.edu",
		dob:             time.Date(2006, 2, 14, 0, 0, 0, 0, time.UTC),
		motherName:      "Kiran Patel",
		fatherName:      "Nirav Patel",
		category:        "ST",
		completeAddress: "12, Gandhi Nagar, Ratlam, MP 457001",
		mobile:          "9876507234",
	},
	{
		studentName:     "Sneha Deshmukh",
		email:           "sneha.deshmukh@example.edu",
		dob:             time.Date(2005, 8, 25, 0, 0, 0, 0, time.UTC),
		motherName:      "Swati Deshmukh",
		fatherName:      "Amit Deshmukh",
		category:        "General",
		completeAddress: "28, Civil Lines, Sagar, MP 470001",
		mobile:          "9876508234",
	},
	{
		studentName:     "Rohan Joshi",
		email:           "rohan.joshi@example.edu",
		dob:             time.Date(2004, 12, 3, 0, 0, 0, 0, time.UTC),
		motherName:      "Meera Joshi",
		fatherName:      "Manish Joshi",
		category:        "EWS",
		completeAddress: "6, Shanti Niketan, Dewas, MP 455001",
		mobile:          "9876509234",
	},
	{
		studentName:     "Neha Verma",
		email:           "neha.verma@example.edu",
		dob:             time.Date(2005, 6, 19, 0, 0, 0, 0, time.UTC),
		motherName:      "Rekha Verma",
		fatherName:      "Pankaj Verma",
		category:        "OBC",
		completeAddress: "14, Ring Road, Rewa, MP 486001",
		mobile:          "9876510234",
	},
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
	var files []p_filesystem.VNode
	if err := db.Where("is_directory = ?", false).Find(&files).Error; err != nil {
		return nil, err
	}
	return files, nil
}

func init() {
	lago.RegistryGenerator.Register("studentapplications.Generator", lago.Generator{
		Create: func(db *gorm.DB) error {
			var programs []p_nirmancampus_programs.Program
			if err := db.Find(&programs).Error; err != nil {
				return fmt.Errorf("load programs: %w", err)
			}
			if len(programs) == 0 {
				return fmt.Errorf("studentapplications.Generator: no programs in database; run programs.Generator first")
			}

			files, err := loadFileNodes(db)
			if err != nil {
				return err
			}

			var seedCreator p_users.User
			var seedCreatorID *uint
			if err := db.First(&seedCreator).Error; err == nil {
				id := seedCreator.ID
				seedCreatorID = &id
			}

			for i := range sampleApplicationRows {
				row := sampleApplicationRows[i]
				prog := programs[i%len(programs)]
				dob := row.dob

				app := StudentApplication{
					ProgramID:       prog.ID,
					CreatedByID:     seedCreatorID,
					StudentName:     row.studentName,
					Email:           row.email,
					DOB:             &dob,
					MotherName:      row.motherName,
					FatherName:      row.fatherName,
					Category:        row.category,
					CompleteAddress: row.completeAddress,
					Mobile:          row.mobile,
				}

				if len(files) > 0 && rand.Intn(100) < 80 {
					picked := files[rand.Intn(len(files))]
					app.PhotoID = &picked.ID
				} else if node, genErr := p_filesystem.GeneratePhotoFile(db); genErr != nil {
					return fmt.Errorf("generate photo for application %q: %w", row.studentName, genErr)
				} else if node != nil {
					app.PhotoID = &node.ID
					files, err = loadFileNodes(db)
					if err != nil {
						return err
					}
				}

				if err := db.Create(&app).Error; err != nil {
					return fmt.Errorf("failed to create student application %q: %w", row.studentName, err)
				}

				nDocs := rand.Intn(4)
				if nDocs == 0 {
					continue
				}
				files, err = loadFileNodes(db)
				if err != nil {
					return err
				}
				docs := pickDistinctFiles(files, nDocs, app.PhotoID)
				if len(docs) == 0 {
					continue
				}
				if err := db.Model(&app).Association("Documents").Append(docs); err != nil {
					return fmt.Errorf("attach documents for %q: %w", row.studentName, err)
				}
			}

			fmt.Printf("Created %d student applications\n", len(sampleApplicationRows))
			return nil
		},
		Remove: func(db *gorm.DB) error {
			if err := db.Exec("DELETE FROM student_application_documents").Error; err != nil {
				slog.Error("failed clearing student_application_documents join table", "error", err)
			}
			return db.Unscoped().Where("1=1").Delete(&StudentApplication{}).Error
		},
	})
}
