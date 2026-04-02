package p_nirmancampus_academicrecords

import (
	"context"
	"fmt"

	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_nirmancampus_courses"
	"github.com/lariv-in/lago/plugins/p_nirmancampus_programs"
	"github.com/lariv-in/lago/plugins/p_nirmancampus_students"
	"gorm.io/gorm"
)

var sampleStatuses = []string{
	"Enrolled",
	"Enrolled",
	"Completed",
	"Probation",
	"Withdrawn",
}

var sampleTerms = []uint{1, 2, 1, 3, 2, 1, 2, 3}

func init() {
	lago.RegistryGenerator.Register("academicrecords.Generator", lago.Generator{
		Create: func(db *gorm.DB) error {
			students, err := gorm.G[p_nirmancampus_students.Student](db).Order("id ASC").Find(context.Background())
			if err != nil {
				return fmt.Errorf("failed to load students: %w", err)
			}
			if len(students) == 0 {
				return fmt.Errorf("need at least one student before generating academic records")
			}

			programs, err := gorm.G[p_nirmancampus_programs.Program](db).Order("id ASC").Find(context.Background())
			if err != nil {
				return fmt.Errorf("failed to load programs: %w", err)
			}
			if len(programs) == 0 {
				return fmt.Errorf("need at least one program before generating academic records")
			}

			courses, err := gorm.G[p_nirmancampus_courses.Course](db).Order("id ASC").Find(context.Background())
			if err != nil {
				return fmt.Errorf("failed to load courses: %w", err)
			}
			if len(courses) == 0 {
				return fmt.Errorf("need at least one course before generating academic records")
			}

			n := 0
			for k, st := range students {
				rec := AcademicRecord{
					StudentID: st.ID,
					ProgramID: programs[k%len(programs)].ID,
					Term:      sampleTerms[k%len(sampleTerms)],
					Status:    sampleStatuses[k%len(sampleStatuses)],
				}
				if err := gorm.G[AcademicRecord](db).Create(context.Background(), &rec); err != nil {
					return fmt.Errorf("failed to create academic record (student_id=%d): %w", st.ID, err)
				}

				primary := courses[k%len(courses)]
				compulsory := []p_nirmancampus_courses.Course{primary}
				var optional []p_nirmancampus_courses.Course
				if len(courses) > 1 {
					optional = append(optional, courses[(k+1)%len(courses)])
				}
				if err := db.Model(&rec).Association("CompulsoryCourses").Append(compulsory); err != nil {
					return fmt.Errorf("failed to attach compulsory courses to academic record (id=%d): %w", rec.ID, err)
				}
				if len(optional) > 0 {
					if err := db.Model(&rec).Association("OptionalCourses").Append(optional); err != nil {
						return fmt.Errorf("failed to attach optional courses to academic record (id=%d): %w", rec.ID, err)
					}
				}

				n++
			}

			fmt.Printf("Created %d academic records (%d students)\n", n, len(students))
			return nil
		},
		Remove: func(db *gorm.DB) error {
			if err := db.Exec("DELETE FROM academic_record_compulsory_courses").Error; err != nil {
				return fmt.Errorf("clear academic_record_compulsory_courses: %w", err)
			}
			if err := db.Exec("DELETE FROM academic_record_optional_courses").Error; err != nil {
				return fmt.Errorf("clear academic_record_optional_courses: %w", err)
			}
			// Legacy join table from the previous single Courses association (safe if absent).
			_ = db.Exec("DELETE FROM academic_record_courses").Error
			return db.Unscoped().Where("1=1").Delete(&AcademicRecord{}).Error
		},
	})
}
