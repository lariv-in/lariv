package p_nirmancampus_programs

import (
	"fmt"

	"github.com/lariv-in/lago/lago"
	"gorm.io/gorm"
)

var samplePrograms = []Program{
	{
		Name:        "B.Sc. Computer Science",
		Code:        "GEN-BSC-CS",
		Description: "Four-year undergraduate program in computer science and software engineering.",
		University:  "IGNOU",
		ProgramType: "bachelor",
	},
	{
		Name:        "B.A. English Literature",
		Code:        "GEN-BA-ENG",
		Description: "Undergraduate degree focusing on literature, composition, and critical analysis.",
		University:  "MRSPTU",
		ProgramType: "bachelor",
	},
	{
		Name:        "B.Com. General",
		Code:        "GEN-BCOM-GEN",
		Description: "Bachelor of Commerce with core accounting, finance, and business foundations.",
		University:  "IGNOU",
		ProgramType: "bachelor",
	},
	{
		Name:        "B.Sc. Physics",
		Code:        "GEN-BSC-PHY",
		Description: "Science program covering classical and modern physics with laboratory work.",
		University:  "MRSPTU",
		ProgramType: "bachelor",
	},
	{
		Name:        "Diploma in Elementary Education",
		Code:        "GEN-DEL-ED",
		Description: "Two-year diploma preparing educators for primary-level teaching.",
		University:  "IGNOU",
		ProgramType: "diploma",
	},
	{
		Name:        "M.A. Political Science",
		Code:        "GEN-MA-POL",
		Description: "Postgraduate program in political theory, public policy, and governance.",
		University:  "MRSPTU",
		ProgramType: "masters",
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
