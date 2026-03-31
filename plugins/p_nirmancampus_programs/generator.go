package p_nirmancampus_programs

import (
	"fmt"

	"github.com/lariv-in/lago/lago"
	courses "github.com/lariv-in/lago/plugins/p_nirmancampus_courses"
	"gorm.io/gorm"
)

// programStructureUnitSeed describes one term; course codes must exist (courses generator runs first).
type programStructureUnitSeed struct {
	TermNumber          int
	OptionalCourseCount int
	CompulsoryCodes     []string
	OptionalPoolCodes   []string
}

// Sample course codes must match p_nirmancampus_courses sampleCourses (run courses generator before programs).
var sampleProgramSeeds = []struct {
	Program Program
	Units   []programStructureUnitSeed
}{
	{
		Program: Program{
			Name:              "B.Sc. Computer Science",
			Code:              "GEN-BSC-CS",
			Description:       "Four-year undergraduate program in computer science and software engineering.",
			University:        "IGNOU",
			ProgramType:       "bachelor",
			AdmissionSessions: AdmissionSessionBoth,
			TermType:          TermTypeSemester,
		},
		Units: []programStructureUnitSeed{
			{TermNumber: 1, OptionalCourseCount: 1, CompulsoryCodes: []string{"GEN-CS101"}, OptionalPoolCodes: []string{"GEN-MTH101", "GEN-ENG205"}},
			{TermNumber: 2, OptionalCourseCount: 0, CompulsoryCodes: []string{"GEN-CS201"}, OptionalPoolCodes: nil},
			{TermNumber: 3, OptionalCourseCount: 1, CompulsoryCodes: []string{"GEN-CS301"}, OptionalPoolCodes: []string{"GEN-CHM201", "GEN-PHY301"}},
		},
	},
	{
		Program: Program{
			Name:              "B.A. English Literature",
			Code:              "GEN-BA-ENG",
			Description:       "Undergraduate degree focusing on literature, composition, and critical analysis.",
			University:        "MRSPTU",
			ProgramType:       "bachelor",
			AdmissionSessions: AdmissionSessionBoth,
			TermType:          TermTypeSemester,
		},
		Units: []programStructureUnitSeed{
			{TermNumber: 1, OptionalCourseCount: 1, CompulsoryCodes: []string{"GEN-ENG205"}, OptionalPoolCodes: []string{"GEN-MTH101", "GEN-CS101"}},
			{TermNumber: 2, OptionalCourseCount: 1, CompulsoryCodes: []string{"GEN-MTH101"}, OptionalPoolCodes: []string{"GEN-ENG205", "GEN-ECO201"}},
		},
	},
	{
		Program: Program{
			Name:              "B.Com. General",
			Code:              "GEN-BCOM-GEN",
			Description:       "Bachelor of Commerce with core accounting, finance, and business foundations.",
			University:        "IGNOU",
			ProgramType:       "bachelor",
			AdmissionSessions: AdmissionSessionBoth,
			TermType:          TermTypeYear,
		},
		Units: []programStructureUnitSeed{
			{TermNumber: 1, OptionalCourseCount: 0, CompulsoryCodes: []string{"GEN-MTH101", "GEN-ECO201"}, OptionalPoolCodes: nil},
			{TermNumber: 2, OptionalCourseCount: 1, CompulsoryCodes: []string{"GEN-CS101"}, OptionalPoolCodes: []string{"GEN-CHM201", "GEN-PHY301"}},
		},
	},
	{
		Program: Program{
			Name:              "B.Sc. Physics",
			Code:              "GEN-BSC-PHY",
			Description:       "Science program covering classical and modern physics with laboratory work.",
			University:        "MRSPTU",
			ProgramType:       "bachelor",
			AdmissionSessions: AdmissionSessionJuly,
			TermType:          TermTypeSemester,
		},
		Units: []programStructureUnitSeed{
			{TermNumber: 1, OptionalCourseCount: 0, CompulsoryCodes: []string{"GEN-MTH101", "GEN-PHY301"}, OptionalPoolCodes: nil},
			{TermNumber: 2, OptionalCourseCount: 1, CompulsoryCodes: []string{"GEN-CHM201"}, OptionalPoolCodes: []string{"GEN-CS101", "GEN-CS201"}},
		},
	},
	{
		Program: Program{
			Name:              "Diploma in Elementary Education",
			Code:              "GEN-DEL-ED",
			Description:       "Two-year diploma preparing educators for primary-level teaching.",
			University:        "IGNOU",
			ProgramType:       "diploma",
			AdmissionSessions: AdmissionSessionJan,
			TermType:          TermTypeSemester,
		},
		Units: []programStructureUnitSeed{
			{TermNumber: 1, OptionalCourseCount: 1, CompulsoryCodes: []string{"GEN-MTH101"}, OptionalPoolCodes: []string{"GEN-ENG205"}},
			{TermNumber: 2, OptionalCourseCount: 0, CompulsoryCodes: []string{"GEN-ENG205"}, OptionalPoolCodes: nil},
		},
	},
	{
		Program: Program{
			Name:              "M.A. Political Science",
			Code:              "GEN-MA-POL",
			Description:       "Postgraduate program in political theory, public policy, and governance.",
			University:        "MRSPTU",
			ProgramType:       "masters",
			AdmissionSessions: AdmissionSessionJuly,
			TermType:          TermTypeYear,
		},
		Units: []programStructureUnitSeed{
			{TermNumber: 1, OptionalCourseCount: 1, CompulsoryCodes: []string{"GEN-ECO201", "GEN-ENG205"}, OptionalPoolCodes: []string{"GEN-CS301", "GEN-CHM201"}},
		},
	},
}

func coursesByCodes(db *gorm.DB, codes []string) ([]courses.Course, error) {
	if len(codes) == 0 {
		return nil, nil
	}
	var out []courses.Course
	if err := db.Where("code IN ?", codes).Find(&out).Error; err != nil {
		return nil, err
	}
	if len(out) != len(codes) {
		if len(out) == 0 && len(codes) > 0 {
			return nil, fmt.Errorf("no courses in database for codes %v; run courses.Generator before programs.Generator (see GeneratorOrder in deployment TOML)", codes)
		}
		return nil, fmt.Errorf("expected %d courses by code, found %d (codes %v)", len(codes), len(out), codes)
	}
	return out, nil
}

func init() {
	lago.RegistryGenerator.Register("programs.Generator", lago.Generator{
		Create: func(db *gorm.DB) error {
			for i := range sampleProgramSeeds {
				seed := sampleProgramSeeds[i]
				p := seed.Program
				if err := db.Create(&p).Error; err != nil {
					return fmt.Errorf("failed to create program %q: %w", p.Code, err)
				}
				for j := range seed.Units {
					su := seed.Units[j]
					u := ProgramStructureUnit{
						ProgramID:           p.ID,
						TermNumber:          su.TermNumber,
						OptionalCourseCount: su.OptionalCourseCount,
					}
					if err := db.Create(&u).Error; err != nil {
						return fmt.Errorf("failed to create structure unit for program %q term %d: %w", p.Code, u.TermNumber, err)
					}
					compulsory, err := coursesByCodes(db, su.CompulsoryCodes)
					if err != nil {
						return fmt.Errorf("program %q term %d compulsory courses: %w", p.Code, u.TermNumber, err)
					}
					optional, err := coursesByCodes(db, su.OptionalPoolCodes)
					if err != nil {
						return fmt.Errorf("program %q term %d optional pool: %w", p.Code, u.TermNumber, err)
					}
					if err := db.Model(&u).Association("CompulsoryCourses").Replace(compulsory); err != nil {
						return fmt.Errorf("program %q term %d compulsory association: %w", p.Code, u.TermNumber, err)
					}
					if err := db.Model(&u).Association("OptionalCourseSelectionPool").Replace(optional); err != nil {
						return fmt.Errorf("program %q term %d optional pool association: %w", p.Code, u.TermNumber, err)
					}
				}
			}
			fmt.Printf("Created %d programs\n", len(sampleProgramSeeds))
			return nil
		},
		Remove: func(db *gorm.DB) error {
			if err := db.Exec("DELETE FROM program_structure_unit_compulsory_courses").Error; err != nil {
				return err
			}
			if err := db.Exec("DELETE FROM program_structure_unit_optional_courses").Error; err != nil {
				return err
			}
			if err := db.Unscoped().Where("1 = 1").Delete(&ProgramStructureUnit{}).Error; err != nil {
				return err
			}
			return db.Unscoped().Where("1=1").Delete(&Program{}).Error
		},
	})
}
