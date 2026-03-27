package p_nirmancampus_academicrecords

import (
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

var sampleSemesterOrYear = []string{
	"Semester 1",
	"Semester 2",
	"Year 1",
	"Year 2",
	"Semester 1",
	"Year 2",
	"Semester 3",
	"Year 3",
}

func init() {
	lago.RegistryGenerator.Register("academicrecords.Generator", lago.Generator{
		Create: func(db *gorm.DB) error {
			var students []p_nirmancampus_students.Student
			if err := db.Order("id ASC").Find(&students).Error; err != nil {
				return fmt.Errorf("failed to load students: %w", err)
			}
			if len(students) == 0 {
				return fmt.Errorf("need at least one student before generating academic records")
			}

			var programs []p_nirmancampus_programs.Program
			if err := db.Order("id ASC").Find(&programs).Error; err != nil {
				return fmt.Errorf("failed to load programs: %w", err)
			}
			if len(programs) == 0 {
				return fmt.Errorf("need at least one program before generating academic records")
			}

			var courses []p_nirmancampus_courses.Course
			if err := db.Order("id ASC").Find(&courses).Error; err != nil {
				return fmt.Errorf("failed to load courses: %w", err)
			}
			if len(courses) == 0 {
				return fmt.Errorf("need at least one course before generating academic records")
			}

			n := 0
			for k, st := range students {
				rec := AcademicRecord{
					StudentID:      st.ID,
					ProgramID:      programs[k%len(programs)].ID,
					Status:         sampleStatuses[k%len(sampleStatuses)],
					SemesterOrYear: sampleSemesterOrYear[k%len(sampleSemesterOrYear)],
				}
				if err := db.Create(&rec).Error; err != nil {
					return fmt.Errorf("failed to create academic record (student_id=%d): %w", st.ID, err)
				}

				primary := courses[k%len(courses)]
				toAttach := []p_nirmancampus_courses.Course{primary}
				if len(courses) > 1 {
					toAttach = append(toAttach, courses[(k+1)%len(courses)])
				}
				if err := db.Model(&rec).Association("Courses").Append(toAttach); err != nil {
					return fmt.Errorf("failed to attach courses to academic record (id=%d): %w", rec.ID, err)
				}

				n++
			}

			fmt.Printf("Created %d academic records (%d students)\n", n, len(students))
			return nil
		},
		Remove: func(db *gorm.DB) error {
			if err := db.Exec("DELETE FROM academic_record_courses").Error; err != nil {
				return fmt.Errorf("clear academic_record_courses: %w", err)
			}
			return db.Unscoped().Where("1=1").Delete(&AcademicRecord{}).Error
		},
	})
}
