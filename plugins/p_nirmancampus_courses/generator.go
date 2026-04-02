package p_nirmancampus_courses

import (
	"context"
	"fmt"

	"github.com/lariv-in/lago/lago"
	"gorm.io/gorm"
)

var sampleCourses = []Course{
	{
		Name:        "Introduction to Programming",
		Code:        "GEN-CS101",
		IsActive:    true,
		CourseType:  "Core",
		Description: "Fundamentals of programming using a high-level language and problem-solving.",
	},
	{
		Name:        "Data Structures",
		Code:        "GEN-CS201",
		IsActive:    true,
		CourseType:  "Core",
		Description: "Lists, trees, graphs, hashing, and basic complexity analysis.",
	},
	{
		Name:        "Database Systems",
		Code:        "GEN-CS301",
		IsActive:    true,
		CourseType:  "Core",
		Description: "Relational model, SQL, transactions, and storage internals.",
	},
	{
		Name:        "Modern English Poetry",
		Code:        "GEN-ENG205",
		IsActive:    true,
		CourseType:  "Elective",
		Description: "Survey of major poets and movements from the late nineteenth century onward.",
	},
	{
		Name:        "Calculus I",
		Code:        "GEN-MTH101",
		IsActive:    true,
		CourseType:  "Foundation",
		Description: "Limits, derivatives, and introductory integration with applications.",
	},
	{
		Name:        "Organic Chemistry I",
		Code:        "GEN-CHM201",
		IsActive:    true,
		CourseType:  "Core",
		Description: "Structure, nomenclature, and reactions of organic compounds.",
	},
	{
		Name:        "Classical Mechanics",
		Code:        "GEN-PHY301",
		IsActive:    true,
		CourseType:  "Core",
		Description: "Newtonian mechanics, conservation laws, and rigid-body motion.",
	},
	{
		Name:        "Macroeconomics",
		Code:        "GEN-ECO201",
		IsActive:    false,
		CourseType:  "Elective",
		Description: "National income, fiscal and monetary policy, and growth (inactive sample course).",
	},
}

func init() {
	lago.RegistryGenerator.Register("courses.Generator", lago.Generator{
		Create: func(db *gorm.DB) error {
			for i := range sampleCourses {
				c := sampleCourses[i]
				if err := gorm.G[Course](db).Create(context.Background(), &c); err != nil {
					return fmt.Errorf("failed to create course %q: %w", c.Code, err)
				}
			}
			fmt.Printf("Created %d courses\n", len(sampleCourses))
			return nil
		},
		Remove: func(db *gorm.DB) error {
			return db.Unscoped().Where("1=1").Delete(&Course{}).Error
		},
	})
}
