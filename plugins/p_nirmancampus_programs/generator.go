package p_nirmancampus_programs

import (
	"fmt"

	"github.com/lariv-in/lago/lago"
	"gorm.io/gorm"
)

// Sample course codes must match p_nirmancampus_courses sampleCourses (run courses generator before programs).
var samplePrograms = []Program{
	{
		Name:              "B.Sc. Computer Science",
		Code:              "GEN-BSC-CS",
		Description:       "Four-year undergraduate program in computer science and software engineering.",
		University:        "IGNOU",
		ProgramType:       "bachelor",
		AdmissionSessions: AdmissionSessionBoth,
		TermType:          TermTypeSemester,
		Structure: ProgramStructure{
			{TermNumber: 1, CompulsoryCourses: []string{"GEN-CS101"}, OptionalCourseCount: 1, OptionalCourseSelectionPool: []string{"GEN-MTH101", "GEN-ENG205"}},
			{TermNumber: 2, CompulsoryCourses: []string{"GEN-CS201"}, OptionalCourseCount: 0, OptionalCourseSelectionPool: nil},
			{TermNumber: 3, CompulsoryCourses: []string{"GEN-CS301"}, OptionalCourseCount: 1, OptionalCourseSelectionPool: []string{"GEN-CHM201", "GEN-PHY301"}},
		},
	},
	{
		Name:              "B.A. English Literature",
		Code:              "GEN-BA-ENG",
		Description:       "Undergraduate degree focusing on literature, composition, and critical analysis.",
		University:        "MRSPTU",
		ProgramType:       "bachelor",
		AdmissionSessions: AdmissionSessionBoth,
		TermType:          TermTypeSemester,
		Structure: ProgramStructure{
			{TermNumber: 1, CompulsoryCourses: []string{"GEN-ENG205"}, OptionalCourseCount: 1, OptionalCourseSelectionPool: []string{"GEN-MTH101", "GEN-CS101"}},
			{TermNumber: 2, CompulsoryCourses: []string{"GEN-MTH101"}, OptionalCourseCount: 1, OptionalCourseSelectionPool: []string{"GEN-ENG205", "GEN-ECO201"}},
		},
	},
	{
		Name:              "B.Com. General",
		Code:              "GEN-BCOM-GEN",
		Description:       "Bachelor of Commerce with core accounting, finance, and business foundations.",
		University:        "IGNOU",
		ProgramType:       "bachelor",
		AdmissionSessions: AdmissionSessionBoth,
		TermType:          TermTypeYear,
		Structure: ProgramStructure{
			{TermNumber: 1, CompulsoryCourses: []string{"GEN-MTH101", "GEN-ECO201"}, OptionalCourseCount: 0, OptionalCourseSelectionPool: nil},
			{TermNumber: 2, CompulsoryCourses: []string{"GEN-CS101"}, OptionalCourseCount: 1, OptionalCourseSelectionPool: []string{"GEN-CHM201", "GEN-PHY301"}},
		},
	},
	{
		Name:              "B.Sc. Physics",
		Code:              "GEN-BSC-PHY",
		Description:       "Science program covering classical and modern physics with laboratory work.",
		University:        "MRSPTU",
		ProgramType:       "bachelor",
		AdmissionSessions: AdmissionSessionJuly,
		TermType:          TermTypeSemester,
		Structure: ProgramStructure{
			{TermNumber: 1, CompulsoryCourses: []string{"GEN-MTH101", "GEN-PHY301"}, OptionalCourseCount: 0, OptionalCourseSelectionPool: nil},
			{TermNumber: 2, CompulsoryCourses: []string{"GEN-CHM201"}, OptionalCourseCount: 1, OptionalCourseSelectionPool: []string{"GEN-CS101", "GEN-CS201"}},
		},
	},
	{
		Name:              "Diploma in Elementary Education",
		Code:              "GEN-DEL-ED",
		Description:       "Two-year diploma preparing educators for primary-level teaching.",
		University:        "IGNOU",
		ProgramType:       "diploma",
		AdmissionSessions: AdmissionSessionJan,
		TermType:          TermTypeSemester,
		Structure: ProgramStructure{
			{TermNumber: 1, CompulsoryCourses: []string{"GEN-MTH101"}, OptionalCourseCount: 1, OptionalCourseSelectionPool: []string{"GEN-ENG205"}},
			{TermNumber: 2, CompulsoryCourses: []string{"GEN-ENG205"}, OptionalCourseCount: 0, OptionalCourseSelectionPool: nil},
		},
	},
	{
		Name:              "M.A. Political Science",
		Code:              "GEN-MA-POL",
		Description:       "Postgraduate program in political theory, public policy, and governance.",
		University:        "MRSPTU",
		ProgramType:       "masters",
		AdmissionSessions: AdmissionSessionJuly,
		TermType:          TermTypeYear,
		Structure: ProgramStructure{
			{TermNumber: 1, CompulsoryCourses: []string{"GEN-ECO201", "GEN-ENG205"}, OptionalCourseCount: 1, OptionalCourseSelectionPool: []string{"GEN-CS301", "GEN-CHM201"}},
		},
	},
}

func init() {
	lago.RegistryGenerator.Register("programs.Generator", lago.Generator{
		Create: func(db *gorm.DB) error {
			for i := range samplePrograms {
				p := samplePrograms[i]
				if err := db.Create(&p).Error; err != nil {
					return fmt.Errorf("failed to create program %q: %w", p.Code, err)
				}
			}
			fmt.Printf("Created %d programs\n", len(samplePrograms))
			return nil
		},
		Remove: func(db *gorm.DB) error {
			return db.Unscoped().Where("1=1").Delete(&Program{}).Error
		},
	})
}
