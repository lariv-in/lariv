package p_nirmancampus_assignmentsubmissions

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

var sampleSubmissionRows = []struct {
	title  string
	status string
	max    int
	marks  int
}{
	{title: "Unit Test 1", status: AssignmentSubmissionStatusCreatedKey, max: 20, marks: 16},
	{title: "Midterm Assignment", status: "marked", max: 40, marks: 31},
	{title: "Project Report", status: "uploaded", max: 30, marks: 0},
	{title: "Lab Practical", status: "marked", max: 10, marks: 8},
}

func init() {
	lago.RegistryGenerator.Register("assignmentsubmissions.Generator", lago.Generator{
		Create: func(db *gorm.DB) error {
			academicRecords, err := gorm.G[p_nirmancampus_academicrecords.AcademicRecord](db).Order("id ASC").Find(context.Background())
			if err != nil {
				return fmt.Errorf("load academic records: %w", err)
			}
			if len(academicRecords) == 0 {
				return fmt.Errorf("assignmentsubmissions.Generator: no academic records in database; run academicrecords.Generator first")
			}

			courses, err := gorm.G[p_nirmancampus_courses.Course](db).Order("id ASC").Find(context.Background())
			if err != nil {
				return fmt.Errorf("load courses: %w", err)
			}
			if len(courses) == 0 {
				return fmt.Errorf("assignmentsubmissions.Generator: no courses in database; run courses.Generator first")
			}

			files, err := gorm.G[p_filesystem.VNode](db).Where("is_directory = ?", false).Find(context.Background())
			if err != nil {
				return fmt.Errorf("load filesystem files: %w", err)
			}

			created := 0
			for i, ar := range academicRecords {
				row := sampleSubmissionRows[i%len(sampleSubmissionRows)]
				submission := AssignmentSubmission{
					AssignmentTitle:  fmt.Sprintf("%s #%d", row.title, i+1),
					MaxMarks:         row.max,
					SubmissionStatus: row.status,
					Marks:            row.marks,
					CourseID:         courses[i%len(courses)].ID,
					AcademicRecordID: ar.ID,
				}
				if err := gorm.G[AssignmentSubmission](db).Create(context.Background(), &submission); err != nil {
					return fmt.Errorf("create assignment submission for academic_record_id=%d: %w", ar.ID, err)
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
					if err := db.Model(&submission).Association("Assets").Append(assets); err != nil {
						return fmt.Errorf("attach assets to assignment submission id=%d: %w", submission.ID, err)
					}
				}

				created++
			}

			fmt.Printf("Created %d assignment submissions\n", created)
			return nil
		},
		Remove: func(db *gorm.DB) error {
			if err := db.Exec("DELETE FROM assignment_submission_assets").Error; err != nil {
				slog.Error("failed clearing assignment_submission_assets join table", "error", err)
			}
			return db.Unscoped().Where("1=1").Delete(&AssignmentSubmission{}).Error
		},
	})
}
