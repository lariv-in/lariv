package p_nirmancampus_examregistrations

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"

	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_filesystem"
	"github.com/lariv-in/lago/plugins/p_nirmancampus_academicrecords"
	"github.com/lariv-in/lago/plugins/p_nirmancampus_courses"
	"gorm.io/gorm"
)

var sampleRegistrationRows = []struct {
	title  string
	status string
	fee    uint
}{
	{title: "Midterm Exam", status: ExamRegistrationStatusNotRegisteredKey, fee: 500},
	{title: "Final Exam", status: "registered", fee: 1200},
	{title: "Practical Exam", status: "registered", fee: 300},
	{title: "Supplementary", status: ExamRegistrationStatusNotRegisteredKey, fee: 800},
}

func init() {
	lago.RegistryGenerator.Register("examregistrations.Generator", lago.Generator{
		Create: func(db *gorm.DB) error {
			academicRecords, err := gorm.G[p_nirmancampus_academicrecords.AcademicRecord](db).Order("id ASC").Find(context.Background())
			if err != nil {
				return fmt.Errorf("load academic records: %w", err)
			}
			if len(academicRecords) == 0 {
				return fmt.Errorf("examregistrations.Generator: no academic records in database; run academicrecords.Generator first")
			}

			courses, err := gorm.G[p_nirmancampus_courses.Course](db).Order("id ASC").Find(context.Background())
			if err != nil {
				return fmt.Errorf("load courses: %w", err)
			}
			if len(courses) == 0 {
				return fmt.Errorf("examregistrations.Generator: no courses in database; run courses.Generator first")
			}

			files, err := gorm.G[p_filesystem.VNode](db).Where("is_directory = ?", false).Find(context.Background())
			if err != nil {
				return fmt.Errorf("load filesystem files: %w", err)
			}

			created := 0
			for i, ar := range academicRecords {
				row := sampleRegistrationRows[i%len(sampleRegistrationRows)]
				c := courses[i%len(courses)]
				reg := ExamRegistration{
					ExamTitle:          fmt.Sprintf("%s #%d", row.title, i+1),
					RegistrationStatus: row.status,
					Fee:                row.fee,
					CourseID:           c.ID,
					AcademicRecordID:   ar.ID,
				}
				if reg.Fee == 0 && c.Fee != 0 {
					reg.Fee = c.Fee
				}
				if err := gorm.G[ExamRegistration](db).Create(context.Background(), &reg); err != nil {
					return fmt.Errorf("create exam registration for academic_record_id=%d: %w", ar.ID, err)
				}

				if len(files) == 0 {
					created++
					continue
				}

				assetCount := rand.Intn(3)
				if assetCount == 0 {
					created++
					continue
				}

				assets := make([]p_filesystem.VNode, 0, assetCount)
				order := rand.Perm(len(files))
				for _, idx := range order {
					assets = append(assets, files[idx])
					if len(assets) >= assetCount {
						break
					}
				}

				if len(assets) > 0 {
					if err := db.Model(&reg).Association("Assets").Append(assets); err != nil {
						return fmt.Errorf("attach assets to exam registration id=%d: %w", reg.ID, err)
					}
				}

				created++
			}

			fmt.Printf("Created %d exam registrations\n", created)
			return nil
		},
		Remove: func(db *gorm.DB) error {
			if err := db.Exec("DELETE FROM exam_registration_assets").Error; err != nil {
				slog.Error("failed clearing exam_registration_assets join table", "error", err)
			}
			return db.Unscoped().Where("1=1").Delete(&ExamRegistration{}).Error
		},
	})
}
