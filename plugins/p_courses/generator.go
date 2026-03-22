package p_courses

import (
	"fmt"
	"log/slog"

	"github.com/lariv-in/lago/lago"
	"gorm.io/gorm"
)

var sampleCourses = []Course{
	{
		Name:        "Introduction to Programming",
		Code:        "GEN-CS101",
		IsActive:    true,
		Level:       "UG",
		Subject:     "Computer Science",
		Description: "Fundamentals of programming using a high-level language and problem-solving.",
	},
	{
		Name:        "Data Structures",
		Code:        "GEN-CS201",
		IsActive:    true,
		Level:       "UG",
		Subject:     "Computer Science",
		Description: "Lists, trees, graphs, hashing, and basic complexity analysis.",
	},
	{
		Name:        "Database Systems",
		Code:        "GEN-CS301",
		IsActive:    true,
		Level:       "UG",
		Subject:     "Computer Science",
		Description: "Relational model, SQL, transactions, and storage internals.",
	},
	{
		Name:        "Modern English Poetry",
		Code:        "GEN-ENG205",
		IsActive:    true,
		Level:       "UG",
		Subject:     "English",
		Description: "Survey of major poets and movements from the late nineteenth century onward.",
	},
	{
		Name:        "Calculus I",
		Code:        "GEN-MTH101",
		IsActive:    true,
		Level:       "UG",
		Subject:     "Mathematics",
		Description: "Limits, derivatives, and introductory integration with applications.",
	},
	{
		Name:        "Organic Chemistry I",
		Code:        "GEN-CHM201",
		IsActive:    true,
		Level:       "UG",
		Subject:     "Chemistry",
		Description: "Structure, nomenclature, and reactions of organic compounds.",
	},
	{
		Name:        "Classical Mechanics",
		Code:        "GEN-PHY301",
		IsActive:    true,
		Level:       "UG",
		Subject:     "Physics",
		Description: "Newtonian mechanics, conservation laws, and rigid-body motion.",
	},
	{
		Name:        "Macroeconomics",
		Code:        "GEN-ECO201",
		IsActive:    false,
		Level:       "UG",
		Subject:     "Economics",
		Description: "National income, fiscal and monetary policy, and growth (inactive sample course).",
	},
}

func init() {
	lago.RegistryGenerator.Register("courses.Generator", lago.Generator{
		Create: func(db *gorm.DB) error {
			for i := range sampleCourses {
				c := sampleCourses[i]
				if err := db.Create(&c).Error; err != nil {
					return fmt.Errorf("failed to create course %q: %w", c.Code, err)
				}
			}
			fmt.Printf("Created %d courses\n", len(sampleCourses))
			return nil
		},
		Remove: func(db *gorm.DB) error {
			if err := db.Exec("DELETE FROM course_teachers").Error; err != nil {
				slog.Error("failed clearing course_teachers join table", "error", err)
			}
			return db.Unscoped().Where("1=1").Delete(&Course{}).Error
		},
	})
}
